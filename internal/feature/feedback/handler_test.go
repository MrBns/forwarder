package feedback_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MrBns/forwarder/internal/feature/feedback"
)

type stubService struct {
	submitResult *feedback.Feedback
	submitErr    error
	listResult   []*feedback.Feedback
	listErr      error
}

func (s *stubService) Submit(_ context.Context, origin string, fields map[string]string) (*feedback.Feedback, error) {
	return s.submitResult, s.submitErr
}
func (s *stubService) List(_ context.Context, _, _ int) ([]*feedback.Feedback, error) {
	return s.listResult, s.listErr
}

func postFeedback(t *testing.T, h *feedback.Handler, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	h.Submit(rr, req)
	return rr
}

func TestHandler_Submit_Created(t *testing.T) {
	svc := &stubService{submitResult: &feedback.Feedback{
		ID: "abc-123", Fields: map[string]string{"msg": "hi"},
		Origin: "https://example.com", CreatedAt: time.Now(),
	}}
	rr := postFeedback(t, feedback.NewHandler(svc), map[string]any{"fields": map[string]string{"msg": "hi"}})
	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	var got feedback.Feedback
	_ = json.NewDecoder(rr.Body).Decode(&got)
	if got.ID != "abc-123" {
		t.Errorf("ID = %q, want abc-123", got.ID)
	}
}

func TestHandler_Submit_EmptyFields(t *testing.T) {
	rr := postFeedback(t, feedback.NewHandler(&stubService{}), map[string]any{"fields": map[string]string{}})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandler_Submit_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	feedback.NewHandler(&stubService{}).Submit(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandler_Submit_ServiceError(t *testing.T) {
	svc := &stubService{submitErr: &fakeErr{"db error"}}
	rr := postFeedback(t, feedback.NewHandler(svc), map[string]any{"fields": map[string]string{"k": "v"}})
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestHandler_List_OK(t *testing.T) {
	svc := &stubService{listResult: []*feedback.Feedback{
		{ID: "1", Fields: map[string]string{"x": "y"}, CreatedAt: time.Now()},
	}}
	req := httptest.NewRequest(http.MethodGet, "/api/feedback?limit=5&offset=0", nil)
	rr := httptest.NewRecorder()
	feedback.NewHandler(svc).List(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var body map[string]any
	_ = json.NewDecoder(rr.Body).Decode(&body)
	items, _ := body["feedbacks"].([]any)
	if len(items) != 1 {
		t.Errorf("expected 1 feedback, got %v", body["feedbacks"])
	}
}

func TestHandler_List_EmptySlice(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/feedback", nil)
	rr := httptest.NewRecorder()
	feedback.NewHandler(&stubService{}).List(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var body map[string]any
	_ = json.NewDecoder(rr.Body).Decode(&body)
	items, _ := body["feedbacks"].([]any)
	if len(items) != 0 {
		t.Errorf("expected empty feedbacks, got %v", body["feedbacks"])
	}
}

type fakeErr struct{ msg string }

func (e *fakeErr) Error() string { return e.msg }
