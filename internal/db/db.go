// Package db provides the shared PostgreSQL connection pool used by features
// that require persistent storage.
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MrBns/forwarder/internal/config"
)

// NewPool opens a pgxpool connection pool using the supplied DBConfig.
// It returns an error when DATABASE_URL is empty or the database is unreachable.
func NewPool(cfg config.DBConfig) (*pgxpool.Pool, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("db: DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("db: open pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	return pool, nil
}
