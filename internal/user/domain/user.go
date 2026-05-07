// package domain

// import "time"

// type UserStatus int8

// const (
// 	StatusActive    UserStatus = 1
// 	StatusSuspended UserStatus = 2
// 	StatusDeleted   UserStatus = 3
// )

// type User struct {
// 	ID        UserID
// 	Username  string
// 	Email     Email
// 	Status    UserStatus
// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// }

// func (u *User) ToDBModel(passwordHash string) any {
// 	panic("unimplemented")
// }

// func NewUser(username string, email Email) (*User, error) {
// 	if username == "" {
// 		return nil, ErrUsernameRequired
// 	}
// 	if err := email.Validate(); err != nil {
// 		return nil, err
// 	}

//		return &User{
//			ID:        NewUserID(),
//			Username:  username,
//			Email:     email,
//			Status:    StatusActive,
//			CreatedAt: time.Now(),
//			UpdatedAt: time.Now(),
//		}, nil
//	}
package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserStatus int8

const (
	StatusActive    UserStatus = 1
	StatusSuspended UserStatus = 2
	StatusDeleted   UserStatus = 3
)

func (s UserStatus) String() string {
	switch s {
	case StatusActive:
		return "active"
	case StatusSuspended:
		return "suspended"
	case StatusDeleted:
		return "deleted"
	default:
		return "unknown"
	}
}

const (
	minPasswordLen = 8
	maxPasswordLen = 72 // bcrypt 最大输入 72 字节
	minUsernameLen = 4
	maxUsernameLen = 32
)

// User 是用户领域实体。PasswordHash 不会出现在任何对外的 DTO 中。
type User struct {
	ID           UserID
	Username     string
	Email        Email
	PasswordHash string
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser 创建新用户。明文密码立刻被 bcrypt 哈希,绝不存原文。
func NewUser(username string, email Email, plainPassword string) (*User, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if err := email.Validate(); err != nil {
		return nil, err
	}
	if err := validatePassword(plainPassword); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		ID:           NewUserID(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Status:       StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// VerifyPassword 校验明文密码。匹配返回 nil,不匹配返回 ErrInvalidPassword。
func (u *User) VerifyPassword(plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plain)); err != nil {
		return ErrInvalidPassword
	}
	return nil
}

// ChangePassword 修改密码(为后续修改密码功能预留)。
// 必须先验证旧密码——这是领域规则,不依赖上层调用方记得做。
func (u *User) ChangePassword(oldPlain, newPlain string) error {
	if err := u.VerifyPassword(oldPlain); err != nil {
		return err
	}
	if err := validatePassword(newPlain); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPlain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	u.UpdatedAt = time.Now()
	return nil
}

// Suspend / Activate 管理用户状态(为管理员功能预留)。
func (u *User) Suspend() {
	u.Status = StatusSuspended
	u.UpdatedAt = time.Now()
}

func (u *User) Activate() {
	u.Status = StatusActive
	u.UpdatedAt = time.Now()
}

func (u *User) IsActive() bool    { return u.Status == StatusActive }
func (u *User) IsSuspended() bool { return u.Status == StatusSuspended }

func validateUsername(username string) error {
	if username == "" {
		return ErrUsernameRequired
	}
	if len(username) < minUsernameLen || len(username) > maxUsernameLen {
		return ErrUsernameRequired
	}
	return nil
}

func validatePassword(plain string) error {
	if len(plain) < minPasswordLen || len(plain) > maxPasswordLen {
		return ErrInvalidPassword
	}
	return nil
}
