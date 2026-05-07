package domain

import "context"

// UserRepository 是用户仓储接口,由 infrastructure 层实现。
// 所有方法都返回 domain.User(包含 PasswordHash)——上层 DTO 决定哪些字段对外暴露。
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByUserID(ctx context.Context, id UserID) (*User, error)
	GetByEmail(ctx context.Context, email Email) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	UpdatePassword(ctx context.Context, id UserID, passwordHash string) error
	UpdateEmailVerified(ctx context.Context, id UserID, verified bool) error
}
