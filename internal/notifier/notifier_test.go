package notifier_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrBns/forwarder/internal/notifier"
)

func TestNewDiscordNotifier_NilWhenNoURL(t *testing.T) {
	if got := notifier.NewDiscordNotifier("", nil); got != nil {
		t.Fatal("expected nil when webhook URL is empty")
	}
}

func TestDiscordNotifier_Name(t *testing.T) {
	d := notifier.NewDiscordNotifier("https://example.com/webhook", nil)
	if d == nil {
		t.Fatal("expected non-nil notifier")
	}
	if d.Name() != "discord" {
		t.Fatalf("expected name 'discord', got %q", d.Name())
	}
}

func TestDiscordNotifier_Send_Success(t *testing.T) {
	var received map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	d := notifier.NewDiscordNotifier(srv.URL, srv.Client())
	if d == nil {
		t.Fatal("expected non-nil notifier")
	}

	msg := notifier.Message{
		Title:       "Hello",
		Description: "World",
		Note:        "test note",
		Footer:      "footer text",
		Fields:      map[string]string{"key": "value"},
		Attachments: []notifier.Attachment{{Name: "doc", URL: "https://example.com/doc"}},
	}
	if err := d.Send(msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, ok := received["content"]
	if !ok || content == "" {
		t.Fatal("expected non-empty content field")
	}
}

func TestDiscordNotifier_Send_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	d := notifier.NewDiscordNotifier(srv.URL, srv.Client())
	if err := d.Send(notifier.Message{Title: "oops"}); err == nil {
		t.Fatal("expected error on non-2xx response")
	}
}

func TestNewSlackNotifier_NilWhenNoURL(t *testing.T) {
	if got := notifier.NewSlackNotifier("", nil); got != nil {
		t.Fatal("expected nil when webhook URL is empty")
	}
}

func TestSlackNotifier_Name(t *testing.T) {
	s := notifier.NewSlackNotifier("https://example.com/webhook", nil)
	if s == nil {
		t.Fatal("expected non-nil notifier")
	}
	if s.Name() != "slack" {
		t.Fatalf("expected name 'slack', got %q", s.Name())
	}
}

func TestSlackNotifier_Send_Success(t *testing.T) {
	var received map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	s := notifier.NewSlackNotifier(srv.URL, srv.Client())
	msg := notifier.Message{
		Title:       "Slack test",
		Description: "hello from slack",
		Note:        "note here",
		Footer:      "powered by forwarder",
		Fields:      map[string]string{"env": "production"},
		Attachments: []notifier.Attachment{{Name: "link", URL: "https://example.com"}},
	}
	if err := s.Send(msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	blocks, ok := received["blocks"]
	if !ok {
		t.Fatal("expected 'blocks' in payload")
	}
	if len(blocks.([]interface{})) == 0 {
		t.Fatal("expected at least one block")
	}
}

func TestSlackNotifier_Send_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	s := notifier.NewSlackNotifier(srv.URL, srv.Client())
	if err := s.Send(notifier.Message{Title: "oops"}); err == nil {
		t.Fatal("expected error on non-2xx response")
	}
}

func TestNewTelegramNotifier_NilWhenMissing(t *testing.T) {
	tests := []struct {
		token  string
		chatID string
	}{
		{"", "123"},
		{"token", ""},
		{"", ""},
	}
	for _, tc := range tests {
		if got := notifier.NewTelegramNotifier(tc.token, tc.chatID, nil); got != nil {
			t.Fatalf("expected nil for token=%q chatID=%q", tc.token, tc.chatID)
		}
	}
}

func TestTelegramNotifier_Name(t *testing.T) {
	tg := notifier.NewTelegramNotifier("mytoken", "mychatid", nil)
	if tg == nil {
		t.Fatal("expected non-nil notifier")
	}
	if tg.Name() != "telegram" {
		t.Fatalf("expected name 'telegram', got %q", tg.Name())
	}
}

func TestTelegramNotifier_Send_Success(t *testing.T) {
	var received map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	// Override the bot API URL via a trick: we need a custom client that redirects
	// to the test server. Here we use the http.Client with a custom transport.
	tg := notifier.NewTelegramNotifierWithBaseURL("mytoken", "12345", srv.Client(), srv.URL)
	if tg == nil {
		t.Fatal("expected non-nil notifier")
	}

	msg := notifier.Message{
		Title:       "TG test",
		Description: "hello telegram",
		Note:        "important note",
		Footer:      "signed",
		Fields:      map[string]string{"user": "alice"},
		Attachments: []notifier.Attachment{{Name: "ref", URL: "https://example.com/ref"}},
	}
	if err := tg.Send(msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["chat_id"] != "12345" {
		t.Fatalf("expected chat_id '12345', got %q", received["chat_id"])
	}
	if received["text"] == "" {
		t.Fatal("expected non-empty text")
	}
}

func TestTelegramNotifier_Send_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	tg := notifier.NewTelegramNotifierWithBaseURL("bad", "123", srv.Client(), srv.URL)
	if err := tg.Send(notifier.Message{Title: "oops"}); err == nil {
		t.Fatal("expected error on non-2xx response")
	}
}
