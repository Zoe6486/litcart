package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"litcart/internal/user/domain"
)

// LoginLimiter 实现 domain.LoginLimiter,基于 Redis。
//
// 策略:
//   - 每次失败 INCR login_fail:{email},首次失败设 TTL = window。
//   - 计数 >= maxAttempts 时 Allow 返回 false,直到 TTL 自然过期。
//   - 成功登录调 Reset 删 key。
//
// 用 INCR + EXPIRE NX 而不是先 GET 再 SET,保证并发安全(原子)。
type LoginLimiter struct {
	client      *redis.Client
	maxAttempts int64
	window      time.Duration
}

var _ domain.LoginLimiter = (*LoginLimiter)(nil)

// NewLoginLimiter 建限流器。常用参数:5 次失败,锁 15 分钟。
func NewLoginLimiter(client *redis.Client, maxAttempts int, window time.Duration) *LoginLimiter {
	return &LoginLimiter{
		client:      client,
		maxAttempts: int64(maxAttempts),
		window:      window,
	}
}

func (l *LoginLimiter) key(email domain.Email) string {
	return fmt.Sprintf("login_fail:%s", email.String())
}

func (l *LoginLimiter) Allow(ctx context.Context, email domain.Email) (bool, error) {
	v, err := l.client.Get(ctx, l.key(email)).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return true, nil
		}
		return false, fmt.Errorf("limiter get: %w", err)
	}
	return v < l.maxAttempts, nil
}

func (l *LoginLimiter) RecordFailure(ctx context.Context, email domain.Email) error {
	key := l.key(email)
	pipe := l.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	// 只在 key 还没 TTL 时设置(首次失败),后续失败不刷新窗口,
	// 避免攻击者每隔一会儿试一次让锁定永远不解除。
	pipe.Expire(ctx, key, l.window)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("limiter incr: %w", err)
	}
	_ = incr.Val()
	return nil
}

func (l *LoginLimiter) Reset(ctx context.Context, email domain.Email) error {
	if err := l.client.Del(ctx, l.key(email)).Err(); err != nil {
		return fmt.Errorf("limiter reset: %w", err)
	}
	return nil
}
