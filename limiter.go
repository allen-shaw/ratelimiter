package ratelimiter

import (
	"context"
	"math"
	"time"

	"github.com/allen-shaw/ratelimiter/internal/limiter"
	"github.com/redis/go-redis/v9"
)

const Inf = math.MaxInt

// 可以将时间转化为速率
// 例如：每5秒一个，转化为速率就是0.2一秒
func Every(n int, interval time.Duration) float64 {
	if interval <= 0 {
		return Inf
	}
	return float64(n) / interval.Seconds()
}

type Limiter interface {
	// 当使用 Wait 方法消费 Token 时，如果此时桶内 Token 数组不足 (小于 N)，
	// 那么 Wait 方法将会阻塞一段时间，直至 Token 满足条件。如果充足则直接返回。
	Wait(ctx context.Context) error
	WaitN(ctx context.Context, n int) error

	// 截止到某一时刻，目前桶中数目是否至少为 n 个，满足则返回 true，
	// 同时从桶中消费 n 个 token
	// 反之返回不消费 Token，false
	Allow(ctx context.Context) (bool, error)
	AllowN(ctx context.Context, now time.Time, n int) (bool, error)

	// 动态调整速率和桶大小
	SetLimit(limit float64)
	SetBurst(b int)
}

func NewLimiter(name string, rdb *redis.Client, limit float64, burst int) Limiter {
	return limiter.NewLimiter(name, rdb, limit, burst)
}
