package formresponse_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrBns/forwarder/features/formresponse"
)

// --- TelegramNotifier tests -------------------------------------------------

func TestNewTelegramNotifier_NilWhenMissingCredentials(t *testing.T) {
	cases := [][2]string{{"", "123"}, {"tok", ""}, {"", ""}}
	for _, c := range cases {
		if got := formresponse.NewTelegramNotifier(c[0], c[1]); got != nil {
			t.Errorf("expected nil for token=%q chatID=%q", c[0], c[1])
		}
	}
}

func TestTelegramNotifier_Send(t *testing.T) {
	var received map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	n := formresponse.NewTelegramNotifierWithBaseURL("testtoken", "42", srv.URL)
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
	if n.Name() != "telegram" {
		t.Errorf("Name() = %q, want telegram", n.Name())
	}
	if err := n.Send(context.Background(), "hello"); err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	if received["chat_id"] != "42" {
		t.Errorf("chat_id = %q, want 42", received["chat_id"])
	}
	if received["text"] != "hello" {
		t.Errorf("text = %q, want hello", received["text"])
	}
}

// --- DiscordNotifier tests --------------------------------------------------

func TestNewDiscordNotifier_NilWhenMissingURL(t *testing.T) {
	if got := formresponse.NewDiscordNotifier(""); got != nil {
		t.Error("expected nil for empty webhook URL")
	}
}

func TestDiscordNotifier_Send(t *testing.T) {
	var received map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	n := formresponse.NewDiscordNotifier(srv.URL)
	if n.Name() != "discord" {
		t.Errorf("Name() = %q, want discord", n.Name())
	}
	if err := n.Send(context.Background(), "hello"); err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	if received["content"] != "hello" {
		t.Errorf("content = %q, want hello", received["content"])
	}
}

// --- Format helpers tests ---------------------------------------------------

func TestFormatHTML_ContainsOriginAndFields(t *testing.T) {
	msg := formresponse.FormatHTML("https://example.com", map[string]string{"name": "Alice"})
	for _, want := range []string{"https://example.com", "Alice", "name"} {
		if !contains(msg, want) {
			t.Errorf("FormatHTML: message missing %q\n%s", want, msg)
		}
	}
}

func TestFormatMarkdown_ContainsOriginAndFields(t *testing.T) {
	msg := formresponse.FormatMarkdown("https://example.com", map[string]string{"email": "a@b.com"})
	for _, want := range []string{"https://example.com", "a@b.com", "email"} {
		if !contains(msg, want) {
			t.Errorf("FormatMarkdown: message missing %q\n%s", want, msg)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
