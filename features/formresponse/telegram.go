package formresponse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// httpClient is the shared HTTP client used by all outgoing notifier adapters.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// TelegramConfig holds the credentials needed to reach the Telegram Bot API.
type TelegramConfig struct {
	BotToken string
	ChatID   string
	// baseURL is overridable in tests; production code leaves it empty.
	baseURL string
}

// TelegramNotifier is the outgoing adapter that forwards messages via the
// Telegram Bot API.  It implements the Notifier outgoing port.
type TelegramNotifier struct {
	cfg TelegramConfig
}

// newTelegramNotifier is the internal constructor shared by production and test
// code.  Returns nil when either credential is missing (feature disabled).
func newTelegramNotifier(cfg TelegramConfig) *TelegramNotifier {
	if cfg.BotToken == "" || cfg.ChatID == "" {
		return nil
	}
	if cfg.baseURL == "" {
		cfg.baseURL = "https://api.telegram.org"
	}
	return &TelegramNotifier{cfg: cfg}
}

// NewTelegramNotifier creates a production TelegramNotifier.
func NewTelegramNotifier(botToken, chatID string) *TelegramNotifier {
	return newTelegramNotifier(TelegramConfig{BotToken: botToken, ChatID: chatID})
}

// NewTelegramNotifierWithBaseURL creates a TelegramNotifier with a custom base
// URL, useful for tests that spin up a fake Telegram server.
func NewTelegramNotifierWithBaseURL(botToken, chatID, baseURL string) *TelegramNotifier {
	return newTelegramNotifier(TelegramConfig{BotToken: botToken, ChatID: chatID, baseURL: baseURL})
}

func (t *TelegramNotifier) Name() string { return "telegram" }

func (t *TelegramNotifier) Send(ctx context.Context, message string) error {
	url := fmt.Sprintf("%s/bot%s/sendMessage", t.cfg.baseURL, t.cfg.BotToken)

	payload := map[string]string{
		"chat_id":    t.cfg.ChatID,
		"text":       message,
		"parse_mode": "HTML",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram: unexpected status %d", resp.StatusCode)
	}
	return nil
}
