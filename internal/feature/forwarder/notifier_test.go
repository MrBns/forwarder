package forwarder_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrBns/forwarder/internal/feature/forwarder"
)

// ── Discord ──────────────────────────────────────────────────────────────────

func TestNewDiscordNotifier_NilWhenNoURL(t *testing.T) {
	if got := forwarder.NewDiscordNotifier("", nil); got != nil {
		t.Fatal("expected nil when webhook URL is empty")
	}
}

func TestDiscordNotifier_Name(t *testing.T) {
	d := forwarder.NewDiscordNotifier("https://example.com/hook", nil)
	if d.Name() != "discord" {
		t.Fatalf("Name() = %q, want discord", d.Name())
	}
}

func TestDiscordNotifier_Send_Success(t *testing.T) {
	var received map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	d := forwarder.NewDiscordNotifier(srv.URL, srv.Client())
	msg := forwarder.Message{
		Title: "Hello", Description: "World", Note: "n",
		Footer: "f", Fields: map[string]string{"k": "v"},
		Attachments: []forwarder.Attachment{{Name: "doc", URL: "https://example.com/doc"}},
	}
	if err := d.Send(msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["content"] == "" {
		t.Fatal("expected non-empty content field")
	}
}

func TestDiscordNotifier_Send_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	d := forwarder.NewDiscordNotifier(srv.URL, srv.Client())
	if err := d.Send(forwarder.Message{Title: "oops"}); err == nil {
		t.Fatal("expected error on non-2xx response")
	}
}

// ── Slack ─────────────────────────────────────────────────────────────────────

func TestNewSlackNotifier_NilWhenNoURL(t *testing.T) {
	if got := forwarder.NewSlackNotifier("", nil); got != nil {
		t.Fatal("expected nil when webhook URL is empty")
	}
}

func TestSlackNotifier_Name(t *testing.T) {
	s := forwarder.NewSlackNotifier("https://example.com/hook", nil)
	if s.Name() != "slack" {
		t.Fatalf("Name() = %q, want slack", s.Name())
	}
}

func TestSlackNotifier_Send_Success(t *testing.T) {
	var received map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	s := forwarder.NewSlackNotifier(srv.URL, srv.Client())
	msg := forwarder.Message{
		Title: "Slack test", Description: "hello", Note: "note",
		Footer: "powered by forwarder", Fields: map[string]string{"env": "prod"},
		Attachments: []forwarder.Attachment{{Name: "link", URL: "https://example.com"}},
	}
	if err := s.Send(msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	blocks, ok := received["blocks"]
	if !ok {
		t.Fatal("expected 'blocks' in payload")
	}
	if len(blocks.([]any)) == 0 {
		t.Fatal("expected at least one block")
	}
}

func TestSlackNotifier_Send_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()
	s := forwarder.NewSlackNotifier(srv.URL, srv.Client())
	if err := s.Send(forwarder.Message{Title: "oops"}); err == nil {
		t.Fatal("expected error on non-2xx response")
	}
}

// ── Telegram ─────────────────────────────────────────────────────────────────

func TestNewTelegramNotifier_NilWhenMissing(t *testing.T) {
	cases := [][2]string{{"", "123"}, {"tok", ""}, {"", ""}}
	for _, c := range cases {
		if got := forwarder.NewTelegramNotifier(c[0], c[1], nil); got != nil {
			t.Errorf("expected nil for token=%q chatID=%q", c[0], c[1])
		}
	}
}

func TestTelegramNotifier_Name(t *testing.T) {
	tg := forwarder.NewTelegramNotifier("tok", "cid", nil)
	if tg.Name() != "telegram" {
		t.Fatalf("Name() = %q, want telegram", tg.Name())
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

	tg := forwarder.NewTelegramNotifierWithBaseURL("mytoken", "12345", srv.Client(), srv.URL)
	msg := forwarder.Message{
		Title: "TG test", Description: "hello", Note: "imp",
		Footer: "signed", Fields: map[string]string{"user": "alice"},
		Attachments: []forwarder.Attachment{{Name: "ref", URL: "https://example.com/ref"}},
	}
	if err := tg.Send(msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["chat_id"] != "12345" {
		t.Fatalf("chat_id = %q, want 12345", received["chat_id"])
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
	tg := forwarder.NewTelegramNotifierWithBaseURL("bad", "123", srv.Client(), srv.URL)
	if err := tg.Send(forwarder.Message{Title: "oops"}); err == nil {
		t.Fatal("expected error on non-2xx response")
	}
}
