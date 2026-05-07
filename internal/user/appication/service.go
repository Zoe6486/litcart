// package application

// import (
// 	"context"

// 	"golang.org/x/crypto/bcrypt"

// 	"litcart/internal/user/domain"
// 	"litcart/pkg/jwt" // 假设你的 JWT 工具包路径
// )

// type UserService struct {
// 	repo domain.UserRepository
// }

// func NewUserService(repo domain.UserRepository) *UserService {
// 	return &UserService{repo: repo}
// }

// func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
// 	email, err := domain.NewEmail(req.Email)
// 	if err != nil {
// 		return nil, err
// 	}

// 	user, err := domain.NewUser(req.Username, email)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if _, err := s.repo.GetByEmail(ctx, req.Email); err == nil {
// 		return nil, domain.ErrEmailExists
// 	}
// 	if _, err := s.repo.GetByUsername(ctx, req.Username); err == nil {
// 		return nil, domain.ErrUsernameExists
// 	}

// 	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := s.repo.Create(ctx, user, string(hashed)); err != nil {
// 		return nil, err
// 	}

// 	return NewUserResponse(user), nil
// }

// func (s *UserService) Login(ctx context.Context, req LoginRequest) (string, error) {
// 	user, err := s.repo.GetByEmail(ctx, req.Email)
// 	if err != nil {
// 		return "", domain.ErrUserNotFound
// 	}

// 	if user.Status == domain.StatusSuspended {
// 		return "", domain.ErrAccountSuspended
// 	}

// 	// 注意：当前 GetByEmail 返回的 User 不包含密码！
// 	// 实际生产中建议增加 GetByEmailWithPassword 方法，这里简化处理
// 	// 或者修改 GetByEmail 返回带密码的临时模型

// 	return jwt.GenToken(user.ID.Int64(), user.Username)
// }

package application

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"litcart/internal/user/domain"
	"litcart/pkg/jwt"
)

const tokenExpiresIn int64 = 86400 // 24h,后续应改为配置项

// UserService 是用户领域的应用服务。
// logger 通过构造函数注入,而不是用 zap.L() 全局变量——方便测试时替换 NopLogger。
type UserService struct {
	repo   domain.UserRepository
	logger *zap.Logger
}

func NewUserService(repo domain.UserRepository, logger *zap.Logger) *UserService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &UserService{repo: repo, logger: logger}
}

// Register 注册新用户。
// 唯一性完全靠数据库索引保证,不在 service 层做预查询(避免 TOCTOU)。
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
		// 已知的领域错误直接透传(handler 用 errors.Is 判断);未知错误才记 error 日志。
		if errors.Is(err, domain.ErrUsernameExists) ||
			errors.Is(err, domain.ErrEmailExists) ||
			errors.Is(err, domain.ErrDuplicateEntry) {
			return nil, err
		}
		s.logger.Error("create user failed",
			zap.String("username", req.Username),
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return nil, fmt.Errorf("register: %w", err)
	}

	s.logger.Info("user registered",
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username),
	)
	return NewUserResponse(user), nil
}

// Login 登录。
// 用户不存在和密码错误都返回 ErrInvalidPassword,防账户枚举;
// 区分级别的安全日志由 handler 决定怎么记(因为 handler 拥有更多请求上下文)。
func (s *UserService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return nil, domain.ErrInvalidPassword
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidPassword
		}
		s.logger.Error("login: query user failed", zap.Error(err))
		return nil, fmt.Errorf("login: %w", err)
	}

	if user.IsSuspended() {
		return nil, domain.ErrAccountSuspended
	}
	if !user.IsActive() {
		// Deleted 或其他不可登录状态,统一当作凭据错误对外
		return nil, domain.ErrInvalidPassword
	}

	if err := user.VerifyPassword(req.Password); err != nil {
		return nil, err
	}

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
