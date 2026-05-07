package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	application "litcart/internal/user/appication"
	"litcart/internal/user/domain"
	"litcart/pkg/response"
)

const requestIDKey = "request_id"

type UserHandler struct {
	service *application.UserService
	logger  *zap.Logger
}

func NewUserHandler(service *application.UserService, logger *zap.Logger) *UserHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &UserHandler{service: service, logger: logger}
}

func (h *UserHandler) withReqLogger(c *gin.Context) *zap.Logger {
	if rid, ok := c.Get(requestIDKey); ok {
		if s, ok := rid.(string); ok && s != "" {
			return h.logger.With(zap.String("request_id", s))
		}
	}
	return h.logger
}

// SignUp POST /users/signup
func (h *UserHandler) SignUp(c *gin.Context) {
	log := h.withReqLogger(c)

	var req application.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailExists):
			response.Conflict(c, "email already registered")
		case errors.Is(err, domain.ErrUsernameExists):
			response.Conflict(c, "username already registered")
		case errors.Is(err, domain.ErrDuplicateEntry):
			response.Conflict(c, "resource already exists")
		case errors.Is(err, domain.ErrInvalidEmail),
			errors.Is(err, domain.ErrUsernameRequired),
			errors.Is(err, domain.ErrPasswordTooShort),
			errors.Is(err, domain.ErrPasswordTooLong),
			errors.Is(err, domain.ErrPasswordTooWeak):
			response.ValidationError(c, err)
		default:
			log.Error("unexpected error in signup",
				zap.String("username", req.Username),
				zap.Error(err),
			)
			response.InternalError(c, err)
		}
		return
	}

	response.Created(c, gin.H{"message": "registration successful", "user": resp})
}

// Login POST /users/login
func (h *UserHandler) Login(c *gin.Context) {
	log := h.withReqLogger(c)

	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAccountSuspended):
			log.Warn("login attempt on suspended account",
				zap.String("email", req.Email),
				zap.String("ip", c.ClientIP()),
			)
			response.Unauthorized(c, "account suspended")
		case errors.Is(err, domain.ErrTooManyAttempts):
			log.Warn("login locked by limiter",
				zap.String("email", req.Email),
				zap.String("ip", c.ClientIP()),
			)
			response.TooManyRequests(c, "too many attempts, try again later")
		case errors.Is(err, domain.ErrInvalidPassword),
			errors.Is(err, domain.ErrUserNotFound):
			log.Warn("invalid login attempt",
				zap.String("email", req.Email),
				zap.String("ip", c.ClientIP()),
			)
			response.Unauthorized(c, "invalid email or password")
		default:
			log.Error("unexpected error in login",
				zap.String("email", req.Email),
				zap.Error(err),
			)
			response.InternalError(c, err)
		}
		return
	}

	response.OK(c, resp)
}

// VerifyEmail POST /users/verify-email
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	log := h.withReqLogger(c)

	var req application.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		switch {
		case errors.Is(err, domain.ErrTokenInvalid):
			response.BadRequest(c, "token invalid or expired")
		case errors.Is(err, domain.ErrUserNotFound):
			response.BadRequest(c, "token invalid or expired")
		default:
			log.Error("verify email failed", zap.Error(err))
			response.InternalError(c, err)
		}
		return
	}
	response.OK(c, gin.H{"message": "email verified"})
}

// ResendVerify POST /users/resend-verify
// 安全:无论结果如何都返回 200,防止被探测哪些 email 已注册。
func (h *UserHandler) ResendVerify(c *gin.Context) {
	log := h.withReqLogger(c)

	var req application.ResendVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.ResendVerifyEmail(c.Request.Context(), req.Email); err != nil {
		log.Error("resend verify failed", zap.Error(err))
	}
	response.OK(c, gin.H{"message": "if the email is registered and unverified, a verification email has been sent"})
}

// ForgotPassword POST /users/forgot-password
// 安全:无论 email 是否注册都返回相同信息。
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	log := h.withReqLogger(c)

	var req application.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		log.Error("forgot password failed", zap.Error(err))
	}
	response.OK(c, gin.H{"message": "if the email is registered, a reset link has been sent"})
}

// ResetPassword POST /users/reset-password
func (h *UserHandler) ResetPassword(c *gin.Context) {
	log := h.withReqLogger(c)

	var req application.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, domain.ErrTokenInvalid):
			response.BadRequest(c, "token invalid or expired")
		case errors.Is(err, domain.ErrPasswordTooShort),
			errors.Is(err, domain.ErrPasswordTooLong),
			errors.Is(err, domain.ErrPasswordTooWeak):
			response.ValidationError(c, err)
		default:
			log.Error("reset password failed", zap.Error(err))
			response.InternalError(c, err)
		}
		return
	}
	response.OK(c, gin.H{"message": "password reset successful"})
}
