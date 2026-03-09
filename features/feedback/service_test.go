package feedback_test

import (
	"context"
	"errors"
	"testing"

	"github.com/MrBns/form-response/features/feedback"
)

// mockRepository is a test double for feedback.Repository.
type mockRepository struct {
	saved  []*feedback.Feedback
	listed []*feedback.Feedback
	saveErr error
	listErr error
}

func (m *mockRepository) Save(_ context.Context, fb *feedback.Feedback) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, fb)
	return nil
}

func (m *mockRepository) List(_ context.Context, _, _ int) ([]*feedback.Feedback, error) {
	return m.listed, m.listErr
}

func TestService_Submit_Success(t *testing.T) {
	repo := &mockRepository{}
	svc := feedback.NewService(repo)

	fb, err := svc.Submit(context.Background(), "https://example.com", map[string]string{
		"name":    "Alice",
		"message": "Great product!",
	})
	if err != nil {
		t.Fatalf("Submit() error: %v", err)
	}
	if fb.ID == "" {
		t.Error("expected non-empty ID")
	}
	if fb.Origin != "https://example.com" {
		t.Errorf("Origin = %q, want https://example.com", fb.Origin)
	}
	if fb.Fields["name"] != "Alice" {
		t.Errorf("Fields[name] = %q, want Alice", fb.Fields["name"])
	}
	if fb.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if len(repo.saved) != 1 {
		t.Errorf("repo.Save called %d times, want 1", len(repo.saved))
	}
}

func TestService_Submit_EmptyFields(t *testing.T) {
	svc := feedback.NewService(&mockRepository{})

	_, err := svc.Submit(context.Background(), "", map[string]string{})
	if err == nil {
		t.Fatal("expected error for empty fields, got nil")
	}
}

func TestService_Submit_RepoError(t *testing.T) {
	repo := &mockRepository{saveErr: errors.New("db down")}
	svc := feedback.NewService(repo)

	_, err := svc.Submit(context.Background(), "", map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected error from repository, got nil")
	}
}

func TestService_List_DefaultLimit(t *testing.T) {
	listed := []*feedback.Feedback{{ID: "1"}, {ID: "2"}}
	repo := &mockRepository{listed: listed}
	svc := feedback.NewService(repo)

	result, err := svc.List(context.Background(), 0, 0) // limit=0 → defaults to 20
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("len(result) = %d, want 2", len(result))
	}
}

func TestService_List_RepoError(t *testing.T) {
	repo := &mockRepository{listErr: errors.New("db error")}
	svc := feedback.NewService(repo)

	_, err := svc.List(context.Background(), 10, 0)
	if err == nil {
		t.Fatal("expected error from repository, got nil")
	}
}
