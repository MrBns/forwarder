// Package feedback implements the feedback-collection feature.
//
// Hexagonal Architecture layout:
//
//   - domain.go      — Feedback domain model; Repository outgoing port;
//     Service incoming port
//   - service.go     — application service  (implements Service)
//   - repository.go  — PostgreSQL outgoing adapter  (implements Repository)
//   - handler.go     — HTTP incoming adapter  (POST /api/feedback, GET /api/feedback)
package feedback

import (
	"context"
	"time"
)

// Feedback is the domain model for a single piece of user feedback.
type Feedback struct {
	ID        string            `json:"id"`
	Fields    map[string]string `json:"fields"`
	Origin    string            `json:"origin"`
	CreatedAt time.Time         `json:"created_at"`
}

// Repository is the outgoing port for persisting and retrieving Feedback.
// The PostgreSQL adapter in repository.go satisfies this interface.
type Repository interface {
	Save(ctx context.Context, fb *Feedback) error
	List(ctx context.Context, limit, offset int) ([]*Feedback, error)
}

// Service is the incoming application-service port called by the HTTP adapter.
type Service interface {
	Submit(ctx context.Context, origin string, fields map[string]string) (*Feedback, error)
	List(ctx context.Context, limit, offset int) ([]*Feedback, error)
}
