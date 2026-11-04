package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"ponpes-saas/internal/config"
	"ponpes-saas/internal/db"
	"ponpes-saas/internal/httpapi"
	"ponpes-saas/internal/repository"
	"ponpes-saas/internal/storage"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		panic("DATABASE_URL is required for Fiber API")
	}
	pool, err := db.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	server := httpapi.NewServer(cfg, repository.Repository{DB: pool}, storage.New(cfg.DataRoot))
	server.SyncProfilesOnStartup(context.Background())
	server.StartCrawlWorker(context.Background())

	// Optional second listener, for the Hermes container only (see deploy/hermes-bridge/README.md). The address is a
	// Docker network gateway that may not exist yet at boot, so keep retrying instead of giving up.
	if listen := strings.TrimSpace(os.Getenv("HERMES_EVENTS_LISTEN")); listen != "" {
		go func() {
			for {
				err := server.EventsApp().Listen(listen)
				slog.Warn("Hermes events listener not running; retrying", "listen", listen, "error", err)
				time.Sleep(5 * time.Second)
			}
		}()
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	if err := server.App().Listen(addr); err != nil {
		panic(err)
	}
}
