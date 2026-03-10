package formresponse_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrBns/forwarder/features/formresponse"
)

// stubNotifier is a test double for formresponse.Notifier.
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
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/submit", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestSubmit_Success(t *testing.T) {
	stub := &stubNotifier{name: "telegram"}
	h := formresponse.NewFormHandler(formresponse.Notifiers{stub})

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
	h := formresponse.NewFormHandler(nil)

	rr := postJSON(t, http.HandlerFunc(h.Submit), map[string]any{"fields": map[string]string{}})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSubmit_InvalidJSON(t *testing.T) {
	h := formresponse.NewFormHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/submit", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Submit(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSubmit_AllNotifiersFail(t *testing.T) {
	stub := &stubNotifier{name: "telegram", sendErr: &testErr{"boom"}}
	h := formresponse.NewFormHandler(formresponse.Notifiers{stub})

	rr := postJSON(t, http.HandlerFunc(h.Submit), map[string]any{
		"fields": map[string]string{"key": "value"},
	})

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestSubmit_PartialFailure_StillReturns200(t *testing.T) {
	ok := &stubNotifier{name: "discord"}
	bad := &stubNotifier{name: "telegram", sendErr: &testErr{"fail"}}
	h := formresponse.NewFormHandler(formresponse.Notifiers{ok, bad})

	rr := postJSON(t, http.HandlerFunc(h.Submit), map[string]any{
		"fields": map[string]string{"key": "value"},
	})

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestSubmit_NoNotifiers_Returns200(t *testing.T) {
	h := formresponse.NewFormHandler(nil)

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
	formresponse.HealthCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

type testErr struct{ msg string }

func (e *testErr) Error() string { return e.msg }
