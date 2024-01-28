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

func TestLimter_NotAllowN(t *testing.T) {
	var (
		name  = "test1"
		rdb   = newRedisClient()
		limit = float64(5)
		burst = 10
	)

	fmt.Println("limit", limit)
	limiter := NewLimiter(name, rdb, limit, burst)
	ctx := context.Background()
	ok, err := limiter.AllowN(ctx, time.Now(), 11)
	assert.Nil(t, err)
	assert.False(t, ok)

	val, err := rdb.Get(ctx, limiter.tokenKey).Int()
	assert.Nil(t, err)
	t.Logf("val: %d", val)
}

func TestTimeNow(t *testing.T) {
	fmt.Println(time.Now().Unix())
}

func TestLimiter_AllowAndWait(t *testing.T) {
	rdb := newRedisClient()
	limiters := make([]*Limiter, 0)
	for range []int{1, 2, 3, 4, 5} {
		limiters = append(limiters, NewLimiter("test", rdb, float64(5), 11))
	}

	okCount := 0
	for _, lim := range limiters {
		ok, err := lim.AllowN(context.Background(), time.Now(), 3)
		if err != nil {
			panic(err)
		}
		if ok {
			okCount++
		}
	}

	assert.Equal(t, 3, okCount)
	assert.Equal(t, 2, redisGet(rdb, "test_token"))

	time.Sleep(2 * time.Second)

	for _, lim := range limiters {
		ok, err := lim.AllowN(context.Background(), time.Now(), 2)
		if err != nil {
			panic(err)
		}
		if ok {
			okCount++
		}
	}

	assert.Equal(t, 8, okCount)
	assert.Equal(t, 1, redisGet(rdb, "test_token"))

}

func TestLimiter_WaitN(t *testing.T) {
	var (
		name  = "test"
		rdb   = newRedisClient()
		limit = float64(5)
		burst = 12
	)

	limiter := NewLimiter(name, rdb, limit, burst)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := limiter.WaitN(ctx, 11)
	assert.Nil(t, err)
	assert.Equal(t, 1, redisGet(rdb, "test_token"))

	startTime := time.Now()

	err = limiter.WaitN(ctx, 11)
	assert.Nil(t, err)
	assert.Equal(t, 1, redisGet(rdb, "test_token"))

	endTime := time.Now()
	waitSec := endTime.Sub(startTime).Seconds()
	fmt.Println("wait_sec", waitSec)
}

func TestLimiter_WaitCancel(t *testing.T) {
	var (
		name  = "test"
		rdb   = newRedisClient()
		limit = float64(5)
		burst = 12
	)

	limiter := NewLimiter(name, rdb, limit, burst)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := limiter.WaitN(ctx, 11)
	assert.Nil(t, err)
	assert.Equal(t, 1, redisGet(rdb, "test_token"))

	err = limiter.WaitN(ctx, 11)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
