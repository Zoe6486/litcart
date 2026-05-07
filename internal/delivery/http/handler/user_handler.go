// package handler

// import (
// 	application "litcart/internal/user/appication"
// 	"litcart/internal/user/domain"

// 	"github.com/gin-gonic/gin"

// 	"litcart/pkg/response"
// )

// type UserHandler struct {
// 	service *application.UserService
// }

// func NewUserHandler(service *application.UserService) *UserHandler {
// 	return &UserHandler{service: service}
// }

// func (h *UserHandler) SignUp(c *gin.Context) {
// 	var req application.CreateUserRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		response.ValidationError(c, err)
// 		return
// 	}

// 	resp, err := h.service.CreateUser(c.Request.Context(), req)
// 	if err != nil {
// 		switch {
// 		case err == domain.ErrEmailExists:
// 			response.Conflict(c, "email already registered")
// 		case err == domain.ErrUsernameExists:
// 			response.Conflict(c, "username already registered")
// 		default:
// 			response.InternalError(c, err)
// 		}
// 		return
// 	}

// 	response.Created(c, gin.H{"message": "registration successful", "user": resp})
// }

// func (h *UserHandler) Login(c *gin.Context) {
// 	var req application.LoginRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		response.ValidationError(c, err)
// 		return
// 	}

// 	token, err := h.service.Login(c.Request.Context(), req)
// 	if err != nil {
// 		response.Unauthorized(c, "invalid email or password")
// 		return
// 	}

//		response.OK(c, gin.H{
//			"access_token": token,
//			"token_type":   "Bearer",
//			"expires_in":   86400,
//		})
//	}
package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	application "litcart/internal/user/appication"
	"litcart/internal/user/domain"
	"litcart/pkg/response"
)

// requestIDKey 是 middleware 注入到 gin.Context 的请求 ID 键。
// 实际值要和 middleware/request_id.go 里设置的一致。
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

// withReqLogger 给 logger 附加 request_id,便于日志追踪。
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
			// 兜底:未识别的唯一索引冲突,不暴露具体字段
			response.Conflict(c, "resource already exists")
		case errors.Is(err, domain.ErrInvalidEmail),
			errors.Is(err, domain.ErrUsernameRequired),
			errors.Is(err, domain.ErrInvalidPassword):
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
		// 安全日志:记录 email + ip 但不返回给用户。统一对外返回防账户枚举。
		switch {
		case errors.Is(err, domain.ErrAccountSuspended):
			log.Warn("login attempt on suspended account",
				zap.String("email", req.Email),
				zap.String("ip", c.ClientIP()),
			)
			response.Unauthorized(c, "account suspended")
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
