package domain

import "unicode"

// ValidatePassword 强度规则:
//  1. 长度 8–72(72 是 bcrypt 的硬上限)
//  2. 至少包含一个字母和一个数字
//
// 这是有意保守的规则。要更严(必须含特殊字符 / 不能含用户名 / 不能在泄露库里),
// 在这里加规则即可,不必动 NewUser / ChangePassword 的调用方。
func ValidatePassword(plain string) error {
	if len(plain) < minPasswordLen {
		return ErrPasswordTooShort
	}
	if len(plain) > maxPasswordLen {
		return ErrPasswordTooLong
	}

	var hasLetter, hasDigit bool
	for _, r := range plain {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
		if hasLetter && hasDigit {
			return nil
		}
	}
	return ErrPasswordTooWeak
}
