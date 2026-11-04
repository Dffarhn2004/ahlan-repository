package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"ponpes-saas/internal/config"
	"ponpes-saas/internal/db"
	"sort"
	"strings"
)

func main() {
	cfg := config.Load()
	pool, err := db.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	ctx := context.Background()
	if _, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		panic(err)
	}
	entries, err := os.ReadDir("db/migrations")
	if err != nil {
		panic(err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		var exists bool
		err = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, entry.Name()).Scan(&exists)
		if err != nil {
			panic(err)
		}
		if exists {
			continue
		}
		data, err := os.ReadFile(filepath.Join("db/migrations", entry.Name()))
		if err != nil {
			panic(err)
		}
		parts := strings.Split(string(data), "-- +goose Down")
		sql := parts[0]
		sql = strings.ReplaceAll(sql, "-- +goose Up", "")
		tx, err := pool.Begin(ctx)
		if err != nil {
			panic(err)
		}
		if _, err = tx.Exec(ctx, sql); err != nil {
			_ = tx.Rollback(ctx)
			panic(fmt.Errorf("%s: %w", entry.Name(), err))
		}
		if _, err = tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, entry.Name()); err != nil {
			_ = tx.Rollback(ctx)
			panic(err)
		}
		if err = tx.Commit(ctx); err != nil {
			panic(err)
		}
		fmt.Println("applied", entry.Name())
	}
}
