package limiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	_ "embed"

	"github.com/redis/go-redis/v9"
)

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
	panic("unimplemented")
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
		return false, fmt.Errorf("run script: %v", err)
	}

	return ok, nil
}
