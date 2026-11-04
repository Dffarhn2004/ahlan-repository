package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"ponpes-saas/internal/config"
	"ponpes-saas/internal/db"
)

func main() {
	cfg := config.Load()
	if cfg.Environment != "development" {
		panic("seed is only allowed when APP_ENV=development")
	}
	email := os.Getenv("SEED_ADMIN_EMAIL")
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	if email == "" {
		email = "admin@localhost"
	}
	if password == "" {
		password = "admin-local-only"
	}
	if len(password) < 10 || strings.Contains(strings.ToLower(password), "password") {
		panic("SEED_ADMIN_PASSWORD must be a non-trivial development password")
	}
	pool, err := db.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	_, err = pool.Exec(context.Background(), `INSERT INTO users(email,password_hash,is_platform_superadmin) VALUES($1,$2,true) ON CONFLICT(email) DO UPDATE SET password_hash=EXCLUDED.password_hash,is_platform_superadmin=true`, email, hash)
	if err != nil {
		panic(err)
	}
	fmt.Printf("seeded development superadmin: %s\n", email)
}
