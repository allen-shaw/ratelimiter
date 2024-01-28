package limiter

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	_ "embed"

	"github.com/redis/go-redis/v9"
)

const InfDuration = time.Duration(math.MaxInt)

var (
	//go:embed script.lua
	script string
)

type Limiter struct {
	mu sync.Mutex

	limit float64
	burst int

	rdb    *redis.Client
	script *redis.Script

	tokenKey     string
	timestampKey string
}

func NewLimiter(name string, rdb *redis.Client, limit float64, burst int) *Limiter {
	return &Limiter{
		limit:        limit,
		burst:        burst,
		rdb:          rdb,
		script:       redis.NewScript(script),
		tokenKey:     name + "_token",
		timestampKey: name + "_ts",
	}
}

// Allow implements ratelimiter.Limiter.
func (l *Limiter) Allow(ctx context.Context) (bool, error) {
	now := time.Now()
	return l.AllowN(ctx, now, 1)
}

// AllowN implements ratelimiter.Limiter.
func (l *Limiter) AllowN(ctx context.Context, now time.Time, n int) (bool, error) {
	return l.reserveN(ctx, now, n)
}

// SetBurst implements ratelimiter.Limiter.
func (l *Limiter) SetBurst(b int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.burst = b
}

// SetLimit implements ratelimiter.Limiter.
func (l *Limiter) SetLimit(limit float64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.limit = limit
}

// Wait implements ratelimiter.Limiter.
func (l *Limiter) Wait(ctx context.Context) error {
	return l.WaitN(ctx, 1)
}

// WaitN implements ratelimiter.Limiter.
func (l *Limiter) WaitN(ctx context.Context, n int) error {
	// 如果 ctx 已经结束了也不用等了
	delay := time.Duration(0)
	timer := time.NewTimer(delay)
	// 等待
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			ok, err := l.reserveN(ctx, time.Now(), n)
			if err != nil {
				return fmt.Errorf("wait reserveN: %v", err)
			}
			if ok {
				return nil
			}
			// fmt.Println("1:", redisGet(l.rdb, l.tokenKey), redisGet(l.rdb, l.timestampKey))
			d := time.Duration(n/int(l.limit)+1) * time.Second
			timer.Reset(d)
		}
	}
}

func (l *Limiter) reserveN(ctx context.Context, now time.Time, n int) (bool, error) {
	ok, err := l.script.Run(ctx,
		l.rdb,
		[]string{l.tokenKey, l.timestampKey},
		l.limit,
		l.burst,
		now.Unix(),
		n,
	).Bool()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("run script: %v", err)
	}

	return ok, nil
}

func redisGet(r *redis.Client, key string) int {
	val, err := r.Get(context.Background(), key).Int()
	if err != nil {
		panic(err)
	}
	return val
}
