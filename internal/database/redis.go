package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedis(ctx context.Context, addr string, password string) (*redis.Client, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	pingCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return rdb, nil
}
