package redisutil

import (
	"context"
	"encoding/json"
	"time"
)

// CacheGet unmarshals a cached value into dest. Returns false if key does not exist.
func CacheGet(ctx context.Context, key string, dest interface{}) bool {
	if Client == nil {
		return false
	}
	data, err := Client.Get(ctx, key).Bytes()
	if err != nil {
		return false
	}
	return json.Unmarshal(data, dest) == nil
}

// CacheSet marshals value to JSON and caches it with a TTL. If ttl is 0, no expiry is set.
func CacheSet(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	if Client == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	Client.Set(ctx, key, data, ttl)
}

// CacheDelete removes one or more keys from Redis. Best-effort.
func CacheDelete(ctx context.Context, keys ...string) {
	if Client == nil || len(keys) == 0 {
		return
	}
	Client.Del(ctx, keys...)
}

// CacheDeletePattern deletes all keys matching a glob pattern. Best-effort.
func CacheDeletePattern(ctx context.Context, pattern string) {
	if Client == nil {
		return
	}
	keys, err := Client.Keys(ctx, pattern).Result()
	if err != nil || len(keys) == 0 {
		return
	}
	Client.Del(ctx, keys...)
}

// CacheExists returns true if the key exists in Redis.
func CacheExists(ctx context.Context, key string) bool {
	if Client == nil {
		return false
	}
	n, err := Client.Exists(ctx, key).Result()
	return err == nil && n > 0
}

// ─── Cache key constants (centralized for easier invalidation) ───

const (
	KeyCredential     = "cache:credential:"   // + appKey
	KeyTemplate       = "cache:template:"     // + id
	KeyStore          = "cache:store:"        // + id
	KeyStatsOverview  = "cache:stats:overview"
	KeyStatsTrend     = "cache:stats:trend:"  // + startDate:endDate
	KeyReportOverview = "cache:report:overview"
	KeyReportTrend    = "cache:report:trend:" // + startDate:endDate
)

// ─── Recommended TTLs ───

const (
	TTLCredential    = 30 * time.Minute
	TTLTemplate      = 10 * time.Minute
	TTLStore         = 10 * time.Minute
	TTLStatsOverview = 60 * time.Second
	TTLStatsTrend    = 5 * time.Minute
	TTLReportOverview = 60 * time.Second
	TTLReportTrend   = 5 * time.Minute
)
