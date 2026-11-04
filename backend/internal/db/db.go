package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	if url == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		slog.Error("database pool creation failed", "error", err)
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		slog.Error("database ping failed", "error", err)
		pool.Close()
		return nil, err
	}
	return pool, nil
}
