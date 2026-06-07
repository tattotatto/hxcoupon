package redisutil

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init(addr, password string, db, poolSize int) error {
	Client = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	return nil
}

func Close() {
	if Client != nil {
		Client.Close()
	}
}

// Available returns true if Redis client is initialized and reachable.
func Available(ctx context.Context) bool {
	if Client == nil {
		return false
	}
	return Client.Ping(ctx).Err() == nil
}
