package redis

import "github.com/gomodule/redigo/redis"

// IncrExpireScript 将incr和expire指令整合成原子操作
var IncrExpireScript = redis.NewScript(1, `
local current
current = redis.call("incr", KEYS[1])
if tonumber(current) == 1 then
    redis.call("expire", KEYS[1], 1)
end`)

// GetDelScript 将get 和del整合成原子操作
var GetDelScript = redis.NewScript(1, `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end`)
