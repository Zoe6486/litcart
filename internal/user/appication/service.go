package application

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"litcart/internal/user/domain"
	"litcart/pkg/jwt"
)

const tokenExpiresIn int64 = 86400 // JWT 过期时间(秒),后续应改为配置项

// UserService 是用户领域的应用服务,装配多个领域端口。
type UserService struct {
	repo    domain.UserRepository
	limiter domain.LoginLimiter
	tokens  domain.MailTokenStore
	mailer  domain.Mailer
	logger  *zap.Logger
}

func NewUserService(
	repo domain.UserRepository,
	limiter domain.LoginLimiter,
	tokens domain.MailTokenStore,
	mailer domain.Mailer,
	logger *zap.Logger,
) *UserService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &UserService{
		repo:    repo,
		limiter: limiter,
		tokens:  tokens,
		mailer:  mailer,
		logger:  logger,
	}
}

// ---------------------------------------------------------------------------
// 注册
// ---------------------------------------------------------------------------

// Register 注册新用户并发出邮箱验证邮件。
// 如果邮件发送失败,用户依然创建成功(可后续手动触发 ResendVerify),
// 不让邮件服务故障影响注册主流程。
func (s *UserService) Register(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return nil, err
	}

	user, err := domain.NewUser(req.Username, email, req.Password)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if isDomainBusinessError(err) {
			return nil, err
		}
		s.logger.Error("create user failed",
			zap.String("username", req.Username),
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return nil, fmt.Errorf("register: %w", err)
	}

	// 发验证邮件——失败不影响注册结果,只记日志
	s.dispatchVerifyEmail(ctx, user)

	s.logger.Info("user registered",
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username),
	)
	return NewUserResponse(user), nil
}

// ---------------------------------------------------------------------------
// 登录
// ---------------------------------------------------------------------------

// Login 登录。
//
// 流程:
//  1. 限流检查(早于查 DB,扛住暴力破解时的 DB 压力)
//  2. 查用户 + 校验状态
//  3. 校验密码,失败计数
//  4. 成功清空计数,签发 JWT
//
// 限流和"用户不存在/密码错"统一返回 ErrInvalidPassword,防账户枚举。
func (s *UserService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return nil, domain.ErrInvalidPassword
	}

	// 1. 限流
	allowed, err := s.limiter.Allow(ctx, email)
	if err != nil {
		// 限流器故障不能阻断登录(否则 Redis 挂了全站登不进),只记日志放行
		s.logger.Error("login limiter check failed", zap.Error(err))
	} else if !allowed {
		return nil, domain.ErrTooManyAttempts
	}

	// 2. 查用户
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// 不存在的 email 也记一次失败,避免攻击者通过响应时间区分用户是否存在
			_ = s.limiter.RecordFailure(ctx, email)
			return nil, domain.ErrInvalidPassword
		}
		s.logger.Error("login: query user failed", zap.Error(err))
		return nil, fmt.Errorf("login: %w", err)
	}

	if user.IsSuspended() {
		return nil, domain.ErrAccountSuspended
	}
	if !user.IsActive() {
		return nil, domain.ErrInvalidPassword
	}

	// 3. 验证密码
	if err := user.VerifyPassword(req.Password); err != nil {
		_ = s.limiter.RecordFailure(ctx, email)
		return nil, err
	}

	// 4. 成功:清失败计数,签 token
	_ = s.limiter.Reset(ctx, email)

	token, err := jwt.GenToken(user.ID.Int64(), user.Username)
	if err != nil {
		s.logger.Error("login: gen token failed",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("login: gen token: %w", err)
	}

	return &LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   tokenExpiresIn,
	}, nil
}

// ---------------------------------------------------------------------------
// 邮箱验证
// ---------------------------------------------------------------------------

// VerifyEmail 校验 token 并把对应用户标记为 email_verified。
func (s *UserService) VerifyEmail(ctx context.Context, token string) error {
	userID, err := s.tokens.Consume(ctx, domain.TokenPurposeEmailVerify, token)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateEmailVerified(ctx, userID, true); err != nil {
		return fmt.Errorf("verify email: %w", err)
	}
	s.logger.Info("email verified", zap.String("user_id", userID.String()))
	return nil
}

// ResendVerifyEmail 重发验证邮件。
// 安全:对"用户不存在"和"已验证"返回相同的 nil,不暴露用户存在性。
func (s *UserService) ResendVerifyEmail(ctx context.Context, emailStr string) error {
	email, err := domain.NewEmail(emailStr)
	if err != nil {
		return nil // 静默忽略,handler 也总是返回 200
	}
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil
		}
		return err
	}
	if user.EmailVerified {
		return nil
	}
	s.dispatchVerifyEmail(ctx, user)
	return nil
}

// ---------------------------------------------------------------------------
// 找回密码
// ---------------------------------------------------------------------------

// ForgotPassword 申请重置密码。
// 安全:无论 email 是否注册都返回 nil,不暴露用户存在性。
func (s *UserService) ForgotPassword(ctx context.Context, emailStr string) error {
	email, err := domain.NewEmail(emailStr)
	if err != nil {
		return nil
	}
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil
		}
		s.logger.Error("forgot password: query failed", zap.Error(err))
		return nil
	}

	token, err := s.tokens.Issue(ctx, domain.TokenPurposePasswordReset, user.ID)
	if err != nil {
		s.logger.Error("forgot password: issue token failed", zap.Error(err))
		return nil
	}
	if err := s.mailer.SendPasswordResetEmail(ctx, user.Email, token); err != nil {
		s.logger.Error("forgot password: send mail failed", zap.Error(err))
	}
	return nil
}

// ResetPassword 用 token 重置密码。
func (s *UserService) ResetPassword(ctx context.Context, token, newPassword string) error {
	userID, err := s.tokens.Consume(ctx, domain.TokenPurposePasswordReset, token)
	if err != nil {
		return err
	}

	user, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	if err := user.ResetPassword(newPassword); err != nil {
		return err
	}
	if err := s.repo.UpdatePassword(ctx, user.ID, user.PasswordHash); err != nil {
		return fmt.Errorf("reset password: %w", err)
	}

	// 重置成功后清掉登录失败计数,避免锁着登不进
	_ = s.limiter.Reset(ctx, user.Email)

	s.logger.Info("password reset", zap.String("user_id", user.ID.String()))
	return nil
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

// dispatchVerifyEmail 发出邮箱验证邮件。失败只记日志,不影响主流程。
func (s *UserService) dispatchVerifyEmail(ctx context.Context, user *domain.User) {
	token, err := s.tokens.Issue(ctx, domain.TokenPurposeEmailVerify, user.ID)
	if err != nil {
		s.logger.Error("issue verify token failed",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return
	}
	if err := s.mailer.SendVerifyEmail(ctx, user.Email, token); err != nil {
		s.logger.Error("send verify mail failed",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
	}
}

// isDomainBusinessError 判断错误是否是已知的业务错误(可以直接透传给 handler)。
func isDomainBusinessError(err error) bool {
	return errors.Is(err, domain.ErrUsernameExists) ||
		errors.Is(err, domain.ErrEmailExists) ||
		errors.Is(err, domain.ErrDuplicateEntry)
}
