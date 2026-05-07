package domain

import "errors"

var (
	// 输入校验
	ErrUsernameRequired = errors.New("username is required")
	ErrEmailRequired    = errors.New("email is required")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrInvalidID        = errors.New("invalid user id")

	// 密码相关
	ErrInvalidPassword  = errors.New("invalid password")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password must be at most 72 characters")
	ErrPasswordTooWeak  = errors.New("password must contain letters and digits")

	// 唯一性冲突
	ErrUsernameExists = errors.New("username already exists")
	ErrEmailExists    = errors.New("email already exists")
	// ErrDuplicateEntry 是唯一索引冲突的兜底,避免泄露原始 MySQL 错误
	ErrDuplicateEntry = errors.New("resource already exists")

	// 状态/查询
	ErrUserNotFound     = errors.New("user not found")
	ErrAccountSuspended = errors.New("account suspended")
	ErrEmailNotVerified = errors.New("email not verified")

	// 限流
	ErrTooManyAttempts = errors.New("too many login attempts, account locked")

	// 邮件 token
	ErrTokenInvalid = errors.New("token invalid or expired")
)
