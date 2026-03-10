// Package feedback implements the feedback-collection feature.
// It follows a Hexagonal Architecture layout:
//
//   - domain.go      — domain model, Repository outgoing port, Service incoming port
//   - service.go     — application service (implements Service)
//   - repository.go  — PostgreSQL outgoing adapter (implements Repository)
//   - handler.go     — HTTP incoming adapter
package feedback

import (
	"context"
	"time"
)

// Feedback is the domain model representing a single piece of feedback.
type Feedback struct {
	ID        string            `json:"id"`
	Fields    map[string]string `json:"fields"`
	Origin    string            `json:"origin"`
	CreatedAt time.Time         `json:"created_at"`
}

// Repository is the outgoing port for persisting and retrieving Feedback.
// The PostgreSQL adapter in repository.go implements this interface.
type Repository interface {
	// Save persists a new Feedback record.
	Save(ctx context.Context, fb *Feedback) error
	// List returns feedbacks ordered by creation time descending.
	List(ctx context.Context, limit, offset int) ([]*Feedback, error)
}

// Service is the incoming application-service port called by the HTTP adapter.
type Service interface {
	// Submit validates, creates, and persists a new Feedback.
	Submit(ctx context.Context, origin string, fields map[string]string) (*Feedback, error)
	// List returns a paginated slice of persisted feedbacks.
	List(ctx context.Context, limit, offset int) ([]*Feedback, error)
}
