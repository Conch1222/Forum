package repository

import (
	"Forum/internal/domain"
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
)

type RateLimiterRepo interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (*domain.RateLimitResult, error)
}

type redisFixedWindowRateLimiter struct {
	rdb    *redis.Client
	script *redis.Script
}

func NewRedisRateLimiter(rdb *redis.Client) RateLimiterRepo {
	// return: allowed(1/0), remaining, retry_after
	lua := `
	local key = KEYS[1]
	local window_ms = tonumber(ARGV[1])
	local limit = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])
	local member = ARGV[4]
	
	local window_start = now - window_ms
	
	-- 1) prune old
	redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)
	
	-- 2) count
	local count = redis.call('ZCARD', key)
	if count >= limit then
	  -- compute retry-after using oldest entry
	  local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
	  local retry_after = 0
	  if #oldest > 0 then
		retry_after = math.ceil((tonumber(oldest[2]) + window_ms - now) / 1000)
	  end
	  return {0, 0, retry_after}
	end
	
	-- 3) add current request
	redis.call('ZADD', key, now, member)
	
	-- 4) expire (keep key slightly longer than window)
	redis.call('PEXPIRE', key, window_ms + 1000)
	
	local remaining = limit - (count + 1)
	if remaining < 0 then remaining = 0 end
	return {1, remaining, 0}
	`

	return &redisFixedWindowRateLimiter{
		rdb:    rdb,
		script: redis.NewScript(lua),
	}
}

func (r *redisFixedWindowRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (*domain.RateLimitResult, error) {
	nowMs := time.Now().UnixMilli()
	member := fmt.Sprintf("%d.%d", nowMs, rand.Int63()) // member must be unique per request

	res, err := r.script.Run(
		ctx, r.rdb, []string{key}, int(window.Milliseconds()),
		limit, nowMs, member).Result()

	if err != nil {
		return nil, err
	}

	arr := res.([]interface{})
	allowed := arr[0].(int64) == 1
	remaining := int(arr[1].(int64))
	retryAfter := int(arr[2].(int64))

	out := &domain.RateLimitResult{
		IsAllowed: allowed,
		Remaining: remaining,
	}

	if !allowed && retryAfter > 0 {
		out.RetryAfter = time.Duration(retryAfter) * time.Second
	}
	return out, nil
}
