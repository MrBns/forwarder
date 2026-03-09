package notifier_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrBns/form-response/internal/notifier"
)

// TestNewTelegramNotifier_NilWhenMissingConfig ensures that a TelegramNotifier
// is not created when either token or chat ID is missing.
func TestNewTelegramNotifier_NilWhenMissingConfig(t *testing.T) {
	tests := []struct {
		token  string
		chatID string
	}{
		{"", "123"},
		{"tok", ""},
		{"", ""},
	}
	for _, tc := range tests {
		if got := notifier.NewTelegramNotifier(tc.token, tc.chatID); got != nil {
			t.Errorf("expected nil for token=%q chatID=%q, got non-nil", tc.token, tc.chatID)
		}
	}
}

// TestNewDiscordNotifier_NilWhenMissingURL ensures that a DiscordNotifier is
// not created when webhookURL is empty.
func TestNewDiscordNotifier_NilWhenMissingURL(t *testing.T) {
	if got := notifier.NewDiscordNotifier(""); got != nil {
		t.Error("expected nil for empty webhook URL, got non-nil")
	}
}

// TestTelegramNotifier_Send exercises the happy path by spinning up a fake
// Telegram Bot API endpoint.
func TestTelegramNotifier_Send(t *testing.T) {
	var received map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	// Monkey-patch the bot API base URL to our test server.
	// We achieve this by using a token that embeds the test server URL through
	// a helper exported solely for tests.
	n := notifier.NewTelegramNotifierWithBaseURL("testtoken", "42", srv.URL)
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
	if n.Name() != "telegram" {
		t.Errorf("name = %q, want %q", n.Name(), "telegram")
	}
	if err := n.Send(context.Background(), "hello"); err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	if received["chat_id"] != "42" {
		t.Errorf("chat_id = %q, want %q", received["chat_id"], "42")
	}
	if received["text"] != "hello" {
		t.Errorf("text = %q, want %q", received["text"], "hello")
	}
}

// TestDiscordNotifier_Send exercises the happy path with a fake Discord webhook.
func TestDiscordNotifier_Send(t *testing.T) {
	var received map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	n := notifier.NewDiscordNotifier(srv.URL)
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
	if n.Name() != "discord" {
		t.Errorf("name = %q, want %q", n.Name(), "discord")
	}
	if err := n.Send(context.Background(), "hello"); err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	if received["content"] != "hello" {
		t.Errorf("content = %q, want %q", received["content"], "hello")
	}
}

// TestFormatFormData checks that the HTML message contains origin and field values.
func TestFormatFormData(t *testing.T) {
	msg := notifier.FormatFormData("https://example.com", map[string]string{
		"name": "Alice",
	})
	if msg == "" {
		t.Fatal("expected non-empty message")
	}
	for _, want := range []string{"https://example.com", "Alice", "name"} {
		if !contains(msg, want) {
			t.Errorf("message does not contain %q:\n%s", want, msg)
		}
	}
}

// TestFormatFormDataPlain checks the plain-text variant used by Discord.
func TestFormatFormDataPlain(t *testing.T) {
	msg := notifier.FormatFormDataPlain("https://example.com", map[string]string{
		"email": "alice@example.com",
	})
	if msg == "" {
		t.Fatal("expected non-empty message")
	}
	for _, want := range []string{"https://example.com", "alice@example.com", "email"} {
		if !contains(msg, want) {
			t.Errorf("message does not contain %q:\n%s", want, msg)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsHelper(s, sub))
}

func containsHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
