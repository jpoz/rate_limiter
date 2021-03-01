package rate

import "github.com/go-redis/redis/v8"

var limiterHeartbeat = redis.NewScript(`
local limit_key = KEYS[1]
local process_id = KEYS[2]

local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

redis.call('ZADD', limit_key, now, process_id)

local clearBefore = now - window
redis.call('ZREMRANGEBYSCORE', limit_key, 0, clearBefore)

redis.call('EXPIRE', limit_key, window)

local process_count = redis.call('ZCARD', limit_key)
return process_count
`)
