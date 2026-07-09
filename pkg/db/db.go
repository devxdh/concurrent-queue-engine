package db

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/001-init.sql
var initSchema string

func Init(databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database configuration string: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to reach target database instance: %w", err)
	}

	return pool, nil
}

func InjectDDL(pool *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("[DB] Injecting DDL...")

	_, err := pool.Exec(ctx, initSchema)
	if err != nil {
		log.Fatalf("[CRITICAL] Failed to Inject DDL: %v", err)
	}

	fmt.Println("[DB] DDL injection successful!")
}
