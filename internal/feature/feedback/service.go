package feedback

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FeedbackService is the application service that implements the Service port.
type FeedbackService struct {
	repo Repository
}

// NewService creates a FeedbackService backed by the supplied Repository.
func NewService(repo Repository) *FeedbackService {
	return &FeedbackService{repo: repo}
}

// Submit validates and persists a new Feedback, returning it on success.
func (s *FeedbackService) Submit(ctx context.Context, origin string, fields map[string]string) (*Feedback, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("feedback: fields must not be empty")
	}
	fb := &Feedback{
		ID:        uuid.New().String(),
		Fields:    fields,
		Origin:    origin,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.Save(ctx, fb); err != nil {
		return nil, fmt.Errorf("feedback: save: %w", err)
	}
	return fb, nil
}

// List returns paginated feedbacks, applying a default limit when limit ≤ 0.
func (s *FeedbackService) List(ctx context.Context, limit, offset int) ([]*Feedback, error) {
	if limit <= 0 {
		limit = 20
	}
	fbs, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("feedback: list: %w", err)
	}
	return fbs, nil
}
