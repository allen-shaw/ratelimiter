package limiter

import (
	"context"
	"log"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestGoRedisLuaScript(t *testing.T) {
	var (
		incrBy = redis.NewScript(`
local key = KEYS[1]
local change = ARGV[1]

local value = redis.call("GET", key)
if not value then
  value = 0
end

value = value + change
redis.call("SET", key, value)

return value
		`)
		sum = redis.NewScript(`
local key = KEYS[1]

local sum = redis.call("GET", key)
if not sum then
	sum = 0
end
		
local num_arg = #ARGV
for i = 1, num_arg do
	sum = sum + ARGV[i]
end
		
redis.call("SET", key, sum)
return sum
		`)
	)

	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	cmd := rdb.FlushDB(ctx)
	if cmd.Err() != nil {
		log.Panicf("redis connect fail: %v", cmd.Err())
	}

	log.Println("# INCR BY")
	for _, change := range []int{1, 5, 0} {
		num, err := incrBy.Run(ctx, rdb, []string{"counter"}, change).Int()
		if err != nil {
			log.Panicf("run incrBy script: %v", err)
		}
		log.Printf("incr by %d: %d", change, num)
	}
	log.Println("# SUM")

	s, err := sum.Run(ctx, rdb, []string{"sum"}, 1, 2, 3).Int()
	if err != nil {
		log.Panicf("run incrBy script: %v", err)
	}
	log.Printf("sum is %d", s)
}
