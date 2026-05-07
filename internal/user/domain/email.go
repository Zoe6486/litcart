package domain

import (
	"regexp"
	"strings"
)

type Email string

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmail 创建 Email 值对象。
// 注意:trim + lowercase 是规范化处理,与数据库 collation 配合保证大小写不敏感。
func NewEmail(email string) (Email, error) {
	e := Email(strings.ToLower(strings.TrimSpace(email)))
	if err := e.Validate(); err != nil {
		return "", err
	}
	return e, nil
}

func (e Email) Validate() error {
	if e == "" {
		return ErrEmailRequired
	}
	if !emailRegex.MatchString(string(e)) {
		return ErrInvalidEmail
	}
	return nil
}

func (e Email) String() string { return string(e) }
