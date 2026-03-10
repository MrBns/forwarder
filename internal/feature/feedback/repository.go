package feedback

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS feedbacks (
    id          TEXT        PRIMARY KEY,
    fields      JSONB       NOT NULL,
    origin      TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

// PostgresRepository is the outgoing PostgreSQL adapter that implements the
// Repository port.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgresRepository and ensures the feedbacks table
// exists.  Returns an error if the auto-migration cannot be applied.
func NewRepository(pool *pgxpool.Pool) (*PostgresRepository, error) {
	if _, err := pool.Exec(context.Background(), createTableSQL); err != nil {
		return nil, fmt.Errorf("feedback: migrate: %w", err)
	}
	return &PostgresRepository{pool: pool}, nil
}

// Save inserts a new Feedback row.
func (r *PostgresRepository) Save(ctx context.Context, fb *Feedback) error {
	fieldsJSON, err := json.Marshal(fb.Fields)
	if err != nil {
		return fmt.Errorf("feedback: marshal fields: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO feedbacks (id, fields, origin, created_at) VALUES ($1, $2, $3, $4)`,
		fb.ID, fieldsJSON, fb.Origin, fb.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("feedback: insert: %w", err)
	}
	return nil
}

// List returns feedbacks ordered newest-first with limit/offset pagination.
func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]*Feedback, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, fields, origin, created_at
		   FROM feedbacks
		  ORDER BY created_at DESC
		  LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("feedback: query: %w", err)
	}
	defer rows.Close()

	var results []*Feedback
	for rows.Next() {
		var fb Feedback
		var fieldsJSON []byte
		var createdAt time.Time
		if err := rows.Scan(&fb.ID, &fieldsJSON, &fb.Origin, &createdAt); err != nil {
			return nil, fmt.Errorf("feedback: scan: %w", err)
		}
		fb.CreatedAt = createdAt
		if err := json.Unmarshal(fieldsJSON, &fb.Fields); err != nil {
			return nil, fmt.Errorf("feedback: unmarshal fields: %w", err)
		}
		results = append(results, &fb)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("feedback: rows: %w", err)
	}
	return results, nil
}
