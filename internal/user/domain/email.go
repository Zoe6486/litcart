package domain

import (
	"regexp"
	"strings"
)

type Email string

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmail 创建并校验 Email 值对象。
func NewEmail(email string) (Email, error) {
	e := Email(strings.TrimSpace(strings.ToLower(email)))
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
