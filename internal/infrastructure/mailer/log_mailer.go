package mailer

import (
	"context"

	"go.uber.org/zap"

	"litcart/internal/user/domain"
)

// LogMailer 是 domain.Mailer 的 mock 实现:把邮件内容打到日志,不真发。
//
// 开发环境用这个,生产替换成 SMTPMailer / SendGridMailer 即可,
// 调用方代码完全不变。
type LogMailer struct {
	logger        *zap.Logger
	verifyURLBase string // 例:https://example.com/verify
	resetURLBase  string // 例:https://example.com/reset
}

var _ domain.Mailer = (*LogMailer)(nil)

func NewLogMailer(logger *zap.Logger, verifyURLBase, resetURLBase string) *LogMailer {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &LogMailer{
		logger:        logger,
		verifyURLBase: verifyURLBase,
		resetURLBase:  resetURLBase,
	}
}

func (m *LogMailer) SendVerifyEmail(ctx context.Context, to domain.Email, token string) error {
	m.logger.Info("[mailer] send verify email",
		zap.String("to", to.String()),
		zap.String("link", m.verifyURLBase+"?token="+token),
	)
	return nil
}

func (m *LogMailer) SendPasswordResetEmail(ctx context.Context, to domain.Email, token string) error {
	m.logger.Info("[mailer] send password reset email",
		zap.String("to", to.String()),
		zap.String("link", m.resetURLBase+"?token="+token),
	)
	return nil
}
