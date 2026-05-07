package domain

import "context"

// Mailer 邮件发送端口。具体实现(SMTP / SendGrid / SES / 日志 Mock)在 infrastructure 层。
type Mailer interface {
	SendVerifyEmail(ctx context.Context, to Email, token string) error
	SendPasswordResetEmail(ctx context.Context, to Email, token string) error
}
