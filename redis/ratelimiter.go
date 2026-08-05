package redis

import (
	"context"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenBucketScript = `
local tokens = redis.call("HGET", KEYS[1], "tokens")
local lastRefill = redis.call("HGET", KEYS[1], "ts")

local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local cost = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

if tokens then
	tokens = tonumber(tokens)
	local elapsed = (now - tonumber(lastRefill)) / 1000
	tokens = math.min(burst, tokens + elapsed * rate)
else
	tokens = burst
end

if tokens >= cost then
	tokens = tokens - cost
	redis.call("HSET", KEYS[1], "tokens", tokens, "ts", now)
	redis.call("PEXPIRE", KEYS[1], tonumber(ARGV[5]))
	return {1, 0}
else
	local waitMs = math.ceil((cost - tokens) / rate * 1000)
	return {0, waitMs}
end
`

type TokenBucket struct {
	client *redis.Client
	script *redis.Script
	rps    float64
	burst  int
	ttlMs  int64
}

func NewTokenBucket(client *redis.Client, rps float64, burst int) *TokenBucket {
	if rps <= 0 {
		panic("ratelimiter: rps must be positive")
	}
	if burst <= 0 {
		panic("ratelimiter: burst must be positive")
	}

	return &TokenBucket{
		client: client,
		script: redis.NewScript(tokenBucketScript),
		rps:    rps,
		burst:  burst,
		ttlMs:  int64(math.Ceil(float64(burst)/rps*1000)) + 60000,
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, time.Duration, error) {
	vals, err := tb.script.Run(
		ctx,
		tb.client,
		[]string{key},
		tb.rps,
		tb.burst,
		1,
		time.Now().UnixMilli(),
		tb.ttlMs,
	).Int64Slice()
	if err != nil {
		return false, 0, err
	}

	return vals[0] == 1, time.Duration(vals[1]) * time.Millisecond, nil
}
