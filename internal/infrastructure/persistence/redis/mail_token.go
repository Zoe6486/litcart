package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"litcart/internal/user/domain"
)

// MailTokenStore 实现 domain.MailTokenStore。
//
// Key 格式:mail_token:{purpose}:{token} → "{userID}"
// purpose 进 key,确保 email_verify 的 token 不能拿来 reset 密码。
type MailTokenStore struct {
	client *redis.Client
	ttl    time.Duration
}

var _ domain.MailTokenStore = (*MailTokenStore)(nil)

// NewMailTokenStore 建 token store。常用 ttl:24h(邮箱验证)/ 1h(密码重置)。
// 这里用同一个 TTL 简化,要细分可以拆成两个实例传不同 ttl,或在接口里加参数。
func NewMailTokenStore(client *redis.Client, ttl time.Duration) *MailTokenStore {
	return &MailTokenStore{client: client, ttl: ttl}
}

func (s *MailTokenStore) key(purpose domain.TokenPurpose, token string) string {
	return fmt.Sprintf("mail_token:%s:%s", purpose, token)
}

func (s *MailTokenStore) Issue(
	ctx context.Context,
	purpose domain.TokenPurpose,
	userID domain.UserID,
) (string, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}
	if err := s.client.Set(ctx, s.key(purpose, token), userID.Int64(), s.ttl).Err(); err != nil {
		return "", fmt.Errorf("issue token: %w", err)
	}
	return token, nil
}

// Consume 用 GETDEL 原子地取出并删除 token,防 token 重放。
func (s *MailTokenStore) Consume(
	ctx context.Context,
	purpose domain.TokenPurpose,
	token string,
) (domain.UserID, error) {
	v, err := s.client.GetDel(ctx, s.key(purpose, token)).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, domain.ErrTokenInvalid
		}
		return 0, fmt.Errorf("consume token: %w", err)
	}
	return domain.UserID(v), nil
}

// randomToken 生成 url-safe 随机 token,n 是字节数(实际字符串更长)。
func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
