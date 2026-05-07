package domain

import "errors"

var (
	ErrUsernameRequired = errors.New("username is required")
	ErrEmailRequired    = errors.New("email is required")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrUsernameExists   = errors.New("username already exists")
	ErrEmailExists      = errors.New("email already exists")
	// ErrDuplicateEntry 是唯一索引冲突的兜底错误。
	// 当未来新增唯一索引(如 phone)、还没在 mapInsertError 里加 case 时,
	// 用户会看到通用的"已存在"提示,而不是原始 MySQL 错误。
	ErrDuplicateEntry   = errors.New("resource already exists")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidID        = errors.New("invalid user id")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrAccountSuspended = errors.New("account suspended")
)
