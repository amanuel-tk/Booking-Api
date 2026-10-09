package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)

	if err != nil {
		return nil, fmt.Errorf("error parsing db url: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)

	if err != nil {
		return nil, fmt.Errorf("error creating pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, time.Second*5)

	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("error pinging db: %w", err)
	}

	return pool, nil

}
