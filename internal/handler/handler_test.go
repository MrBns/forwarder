package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrBns/form-response/internal/handler"
)

// stubNotifier is a test double for notifier.Notifier.
type stubNotifier struct {
	name    string
	sendErr error
	calls   []string
}

func (s *stubNotifier) Name() string { return s.name }
func (s *stubNotifier) Send(_ context.Context, msg string) error {
	s.calls = append(s.calls, msg)
	return s.sendErr
}

func postJSON(t *testing.T, h http.Handler, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/submit", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestSubmit_Success(t *testing.T) {
	stub := &stubNotifier{name: "telegram"}
	h := handler.NewFormHandler(stub)

	rr := postJSON(t, http.HandlerFunc(h.Submit), map[string]any{
		"fields": map[string]string{"name": "Alice", "email": "alice@example.com"},
	})

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if len(stub.calls) != 1 {
		t.Errorf("notifier called %d times, want 1", len(stub.calls))
	}
}

func TestSubmit_EmptyFields(t *testing.T) {
	h := handler.NewFormHandler()

	rr := postJSON(t, http.HandlerFunc(h.Submit), map[string]any{"fields": map[string]string{}})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSubmit_InvalidJSON(t *testing.T) {
	h := handler.NewFormHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/submit", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Submit(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSubmit_AllNotifiersFail(t *testing.T) {
	stub := &stubNotifier{name: "telegram", sendErr: &testError{"boom"}}
	h := handler.NewFormHandler(stub)

	rr := postJSON(t, http.HandlerFunc(h.Submit), map[string]any{
		"fields": map[string]string{"key": "value"},
	})

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestSubmit_PartialNotifierFailure(t *testing.T) {
	// One notifier succeeds, one fails → should still return 200.
	ok := &stubNotifier{name: "discord"}
	bad := &stubNotifier{name: "telegram", sendErr: &testError{"fail"}}
	h := handler.NewFormHandler(ok, bad)

	rr := postJSON(t, http.HandlerFunc(h.Submit), map[string]any{
		"fields": map[string]string{"key": "value"},
	})

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestSubmit_NilNotifiersIgnored(t *testing.T) {
	// Passing nil notifiers must not panic.
	h := handler.NewFormHandler(nil, nil)

	rr := postJSON(t, http.HandlerFunc(h.Submit), map[string]any{
		"fields": map[string]string{"key": "value"},
	})

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.HealthCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
