package application

import (
	"time"

	"litcart/internal/user/domain"
)

// ---- 入参 DTO ----

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=4,max=32"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"` // bcrypt 上限 72 字节
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ---- 响应 DTO ----
//
// UserResponse 不包含 PasswordHash,任何对外暴露的用户信息都走这里。
type UserResponse struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"` // RFC3339 by default
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUserResponse(u *domain.User) *UserResponse {
	return &UserResponse{
		UserID:    u.ID.String(),
		Username:  u.Username,
		Email:     u.Email.String(),
		Status:    u.Status.String(),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}
