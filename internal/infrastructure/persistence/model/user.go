package model

import (
	"time"

	"litcart/internal/user/domain"
)

// User 是数据库行模型。
//
// 时间戳由 MySQL 维护:
//
//	created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//	updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
//
// FromDomain 时若实体里时间戳为零值,这两列让 DB 默认值生效。
type User struct {
	ID            int64     `db:"id"`
	UserID        int64     `db:"user_id"`
	Username      string    `db:"username"`
	Email         string    `db:"email"`
	Password      string    `db:"password"`
	Status        int8      `db:"status"`
	EmailVerified bool      `db:"email_verified"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (m *User) ToDomain() *domain.User {
	email, err := domain.NewEmail(m.Email)
	if err != nil {
		panic("model: corrupted email in db: " + m.Email)
	}
	return &domain.User{
		ID:            domain.UserID(m.UserID),
		Username:      m.Username,
		Email:         email,
		PasswordHash:  m.Password,
		Status:        domain.UserStatus(m.Status),
		EmailVerified: m.EmailVerified,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func FromDomain(u *domain.User) *User {
	return &User{
		UserID:        u.ID.Int64(),
		Username:      u.Username,
		Email:         u.Email.String(),
		Password:      u.PasswordHash,
		Status:        int8(u.Status),
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}
