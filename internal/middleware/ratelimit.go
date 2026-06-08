package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"hxcoupon/internal/dto/response"
	redisutil "hxcoupon/internal/pkg/redis"

	"github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
)

// RateLimit returns a Redis-based sliding window rate limiter middleware.
func RateLimit(keyPrefix string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		storeID, exists := c.Get("store_id")
		if !exists {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		key := fmt.Sprintf("%s:%d", keyPrefix, storeID)

		allowed, err := checkRateLimit(ctx, key, limit, window)
		if err != nil {
			c.Next()
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, response.Error(60008, "rate limit exceeded"))
			return
		}
		c.Next()
	}
}

func checkRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if redisutil.Client == nil {
		return true, nil
	}

	now := time.Now().UnixNano()
	windowStart := now - window.Nanoseconds()

	pipe := redisutil.Client.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	countCmd := pipe.ZCard(ctx, key)
	member := redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)}
	pipe.ZAdd(ctx, key, member)
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return true, err
	}

	count, _ := countCmd.Result()
	return count < int64(limit), nil
}
