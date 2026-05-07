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
	maxPasswordLen = 72
	minUsernameLen = 4
	maxUsernameLen = 32
)

// User 是用户领域实体。
//
// 时间戳设计:
//
//	CreatedAt / UpdatedAt 由数据库 DEFAULT CURRENT_TIMESTAMP 与 ON UPDATE 维护,
//	Go 端不主动赋值。新建实体时这两个字段是零值,Insert 后由 repository 回填。
//	这样多个服务/多副本写同一张表也不会有时钟漂移。
type User struct {
	ID            UserID
	Username      string
	Email         Email
	PasswordHash  string
	Status        UserStatus
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewUser 创建新用户。EmailVerified 默认 false,需要走邮件验证流程激活。
func NewUser(username string, email Email, plainPassword string) (*User, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if err := email.Validate(); err != nil {
		return nil, err
	}
	if err := ValidatePassword(plainPassword); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:            NewUserID(),
		Username:      username,
		Email:         email,
		PasswordHash:  string(hash),
		Status:        StatusActive,
		EmailVerified: false,
		// CreatedAt / UpdatedAt 由 DB 填
	}, nil
}

func (u *User) VerifyPassword(plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plain)); err != nil {
		return ErrInvalidPassword
	}
	return nil
}

// ChangePassword 改密码。强制验证旧密码——领域规则,不依赖上层记得做。
func (u *User) ChangePassword(oldPlain, newPlain string) error {
	if err := u.VerifyPassword(oldPlain); err != nil {
		return err
	}
	return u.ResetPassword(newPlain)
}

// ResetPassword 直接设置新密码(用于"忘记密码"流程,token 已校验过)。
func (u *User) ResetPassword(newPlain string) error {
	if err := ValidatePassword(newPlain); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPlain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) MarkEmailVerified() { u.EmailVerified = true }
func (u *User) Suspend()           { u.Status = StatusSuspended }
func (u *User) Activate()          { u.Status = StatusActive }

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
