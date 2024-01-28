package limiter

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newRedisClient() *redis.Client {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	cmd := rdb.FlushDB(ctx)
	if cmd.Err() != nil {
		log.Panicf("redis connect fail: %v", cmd.Err())
	}
	return rdb
}

func every(n int, interval time.Duration) float64 {
	return float64(n) / interval.Seconds()
}

func TestNewLimiter(t *testing.T) {
	var (
		name  = "test1"
		rdb   = newRedisClient()
		limit = every(30, time.Minute)
		burst = 60
	)

	fmt.Println("limit", limit)
	limiter := NewLimiter(name, rdb, limit, burst)
	ctx := context.Background()
	ok, err := limiter.AllowN(ctx, time.Now(), 10)
	assert.Nil(t, err)
	assert.True(t, ok)

	val, err := rdb.Get(ctx, limiter.tokenKey).Int()
	assert.Nil(t, err)
	t.Logf("val: %d", val)

	ok, err = limiter.AllowN(ctx, time.Now(), 60)
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestLimter_AllowN(t *testing.T) {

}
