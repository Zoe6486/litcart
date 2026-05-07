package domain

import "context"

// LoginLimiter 登录失败限流器。具体实现(Redis/内存)在 infrastructure 层。
//
// 工作模式:
//   每次登录失败调 RecordFailure(email);Allow(email) 返回 false 时拒绝登录尝试;
//   登录成功调 Reset(email) 清掉计数。
type LoginLimiter interface {
	// Allow 返回 true 表示允许尝试,false 表示已被锁定。
	Allow(ctx context.Context, email Email) (bool, error)
	// RecordFailure 记录一次失败。
	RecordFailure(ctx context.Context, email Email) error
	// Reset 清掉失败计数(成功登录时调用)。
	Reset(ctx context.Context, email Email) error
}
