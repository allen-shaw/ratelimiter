local token_key = KEYS[1]
local ts_key = KEYS[2]
local limit = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local n = tonumber(ARGV[4])

local ttl = math.floor((burst/limit)*2)

local cur_tokens = tonumber(redis.call("get", token_key))
if cur_tokens == nil then 
    -- 第一次进入，默认设置当前容量为桶容量
    cur_tokens = burst
end

local last_ts_sec = tonumber(redis.call("get", ts_key))
if last_ts_sec == nil then 
    last_ts_sec = 0
end

-- 距离上次请求的时间间隔
local delta_sec = math.max(0, now-last_ts_sec)
-- 要补充的token 
local filled_token = math.min(burst, cur_tokens + (delta_sec*limit))

local allowed = filled_token >= n

local new_tokens = filled_token
if allowed then 
    new_tokens = new_tokens - n
end 

redis.call("setex", token_key, ttl, new_tokens)
redis.call("setex", ts_key, ttl, now)

return allowed