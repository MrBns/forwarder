package feedback

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FeedbackService is the concrete application service that implements the
// Service incoming port.  It orchestrates between the HTTP adapter and the
// Repository outgoing port.
type FeedbackService struct {
	repo Repository
}

// NewService creates a FeedbackService backed by the supplied Repository.
func NewService(repo Repository) *FeedbackService {
	return &FeedbackService{repo: repo}
}

// Submit creates a new Feedback from the supplied origin and fields, assigns
// it a UUID, and persists it via the Repository.
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

// List delegates to the Repository to return paginated feedbacks.
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
