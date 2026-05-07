package application

import (
	"time"

	"litcart/internal/user/domain"
)

// ---- 注册 / 登录 ----

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=4,max=32"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// ---- 邮件验证 ----

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type ResendVerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ---- 找回密码 ----

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"        binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

// ---- User 视图 ----

type UserResponse struct {
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	Status        string    `json:"status"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func NewUserResponse(u *domain.User) *UserResponse {
	return &UserResponse{
		UserID:        u.ID.String(),
		Username:      u.Username,
		Email:         u.Email.String(),
		Status:        u.Status.String(),
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}
