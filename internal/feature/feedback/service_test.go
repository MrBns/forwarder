package feedback_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrbns/forwarder/internal/feature/feedback"
)

type mockRepository struct {
	saved   []*feedback.Feedback
	listed  []*feedback.Feedback
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
	fb, err := feedback.NewService(repo).Submit(context.Background(), "https://example.com", map[string]string{
		"name": "Alice", "message": "Great product!",
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
	_, err := feedback.NewService(&mockRepository{}).Submit(context.Background(), "", map[string]string{})
	if err == nil {
		t.Fatal("expected error for empty fields")
	}
}

func TestService_Submit_RepoError(t *testing.T) {
	repo := &mockRepository{saveErr: errors.New("db down")}
	_, err := feedback.NewService(repo).Submit(context.Background(), "", map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected error from repository")
	}
}

func TestService_List_DefaultLimit(t *testing.T) {
	repo := &mockRepository{listed: []*feedback.Feedback{{ID: "1"}, {ID: "2"}}}
	result, err := feedback.NewService(repo).List(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("len(result) = %d, want 2", len(result))
	}
}

func TestService_List_RepoError(t *testing.T) {
	repo := &mockRepository{listErr: errors.New("db error")}
	_, err := feedback.NewService(repo).List(context.Background(), 10, 0)
	if err == nil {
		t.Fatal("expected error from repository")
	}
}
