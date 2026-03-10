package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrbns/forwarder/internal/handler"
	"github.com/mrbns/forwarder/internal/notifier"
)

// mockNotifier is a simple in-memory notifier for testing.
type mockNotifier struct {
	name    string
	sendErr error
	sent    []notifier.Message
}

func (m *mockNotifier) Name() string { return m.name }
func (m *mockNotifier) Send(msg notifier.Message) error {
	m.sent = append(m.sent, msg)
	return m.sendErr
}

func post(handler http.Handler, body interface{}) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/forward", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestForwardHandler_AllPlatforms(t *testing.T) {
	discord := &mockNotifier{name: "discord"}
	slack := &mockNotifier{name: "slack"}

	h := handler.NewForwardHandler([]notifier.Notifier{discord, slack})

	rr := post(h, handler.ForwardRequest{
		Title:       "Hello",
		Description: "World",
	})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp handler.ForwardResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resp.Results))
	}
	for _, r := range resp.Results {
		if !r.Success {
			t.Errorf("platform %s failed: %s", r.Platform, r.Error)
		}
	}
}

func TestForwardHandler_FilterByPlatform(t *testing.T) {
	discord := &mockNotifier{name: "discord"}
	slack := &mockNotifier{name: "slack"}
	telegram := &mockNotifier{name: "telegram"}

	h := handler.NewForwardHandler([]notifier.Notifier{discord, slack, telegram})

	rr := post(h, handler.ForwardRequest{
		Title:     "targeted",
		Platforms: []string{"discord", "telegram"},
	})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp handler.ForwardResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resp.Results))
	}
	if len(slack.sent) != 0 {
		t.Fatal("slack should not have received any messages")
	}
}

func TestForwardHandler_NoPlatformsEnabled(t *testing.T) {
	h := handler.NewForwardHandler(nil)

	rr := post(h, handler.ForwardRequest{Title: "test"})

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}

func TestForwardHandler_PlatformNotFound(t *testing.T) {
	discord := &mockNotifier{name: "discord"}
	h := handler.NewForwardHandler([]notifier.Notifier{discord})

	rr := post(h, handler.ForwardRequest{
		Title:     "targeted",
		Platforms: []string{"nonexistent"},
	})

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}

func TestForwardHandler_PartialFailure(t *testing.T) {
	discord := &mockNotifier{name: "discord"}
	slack := &mockNotifier{name: "slack", sendErr: fmt.Errorf("webhook offline")}

	h := handler.NewForwardHandler([]notifier.Notifier{discord, slack})

	rr := post(h, handler.ForwardRequest{Title: "partial"})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp handler.ForwardResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	results := map[string]notifier.Result{}
	for _, r := range resp.Results {
		results[r.Platform] = r
	}

	if !results["discord"].Success {
		t.Error("discord should have succeeded")
	}
	if results["slack"].Success {
		t.Error("slack should have failed")
	}
	if results["slack"].Error == "" {
		t.Error("slack error message should be non-empty")
	}
}

func TestForwardHandler_InvalidJSON(t *testing.T) {
	h := handler.NewForwardHandler([]notifier.Notifier{&mockNotifier{name: "discord"}})

	req := httptest.NewRequest(http.MethodPost, "/api/forward", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestForwardHandler_MessageFields(t *testing.T) {
	mock := &mockNotifier{name: "discord"}
	h := handler.NewForwardHandler([]notifier.Notifier{mock})

	req := handler.ForwardRequest{
		Title:       "title",
		Description: "desc",
		Note:        "note",
		Footer:      "footer",
		Fields:      map[string]string{"k": "v"},
		Attachments: []notifier.Attachment{{Name: "file", URL: "https://example.com/file", Type: "link"}},
	}
	rr := post(h, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if len(mock.sent) != 1 {
		t.Fatalf("expected 1 sent message, got %d", len(mock.sent))
	}

	sent := mock.sent[0]
	if sent.Title != req.Title {
		t.Errorf("title mismatch: got %q", sent.Title)
	}
	if sent.Description != req.Description {
		t.Errorf("description mismatch")
	}
	if sent.Note != req.Note {
		t.Errorf("note mismatch")
	}
	if sent.Footer != req.Footer {
		t.Errorf("footer mismatch")
	}
}
