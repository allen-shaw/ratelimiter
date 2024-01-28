package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/allen-shaw/ratelimiter"
	"github.com/redis/go-redis/v9"
)

func main() {
	var (
		name  = "test"
		rdb   = newRedisClient()
		limit = float64(5)
		burst = 12
	)

	limiter := ratelimiter.NewLimiter(name, rdb, limit, burst)
	ctx := context.Background()
	ok, err := limiter.AllowN(ctx, time.Now(), 10)
	if err != nil {
		panic(err)
	}
	if !ok {
		return
	}
	// do something
	fmt.Println("do something")
}

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
