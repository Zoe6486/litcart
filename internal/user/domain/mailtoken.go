package domain

import "context"

// TokenPurpose 区分 token 用途,避免一个 token 跨场景使用。
type TokenPurpose string

const (
	TokenPurposeEmailVerify   TokenPurpose = "email_verify"
	TokenPurposePasswordReset TokenPurpose = "password_reset"
)

// MailTokenStore 邮件 token 存储。
//
// 设计要点:
//   - 一次性:Consume 成功后立即删除,防 token 重放。
//   - 带过期:实现层(Redis)用 TTL 自动清理。
//   - 绑定用户:token 解析出 UserID,不直接信任客户端传的 ID。
type MailTokenStore interface {
	// Issue 生成并存储一个 token,返回明文 token(发给用户)。
	Issue(ctx context.Context, purpose TokenPurpose, userID UserID) (string, error)
	// Consume 校验并消费 token。成功返回绑定的 UserID,失败返回 ErrTokenInvalid。
	Consume(ctx context.Context, purpose TokenPurpose, token string) (UserID, error)
}
