package submitter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// sharedClient is the HTTP client shared by all submitter outgoing adapters.
var sharedClient = &http.Client{Timeout: 10 * time.Second}

// telegramCfg bundles the credentials for the Telegram Bot API.
type telegramCfg struct {
	botToken string
	chatID   string
	baseURL  string
}

// TelegramNotifier is the outgoing adapter that forwards simple string messages
// via the Telegram Bot API using HTML parse mode.
// It implements the Notifier outgoing port.
type TelegramNotifier struct {
	cfg telegramCfg
}

// NewTelegramNotifier creates a production TelegramNotifier.
// Returns nil when either credential is missing (adapter disabled).
func NewTelegramNotifier(botToken, chatID string) *TelegramNotifier {
	return newTelegramNotifier(telegramCfg{botToken: botToken, chatID: chatID})
}

// NewTelegramNotifierWithBaseURL creates a TelegramNotifier with a custom API
// base URL — intended for use in tests that spin up a fake server.
func NewTelegramNotifierWithBaseURL(botToken, chatID, baseURL string) *TelegramNotifier {
	return newTelegramNotifier(telegramCfg{botToken: botToken, chatID: chatID, baseURL: baseURL})
}

func newTelegramNotifier(cfg telegramCfg) *TelegramNotifier {
	if cfg.botToken == "" || cfg.chatID == "" {
		return nil
	}
	if cfg.baseURL == "" {
		cfg.baseURL = "https://api.telegram.org"
	}
	return &TelegramNotifier{cfg: cfg}
}

func (t *TelegramNotifier) Name() string { return "telegram" }

func (t *TelegramNotifier) Send(ctx context.Context, message string) error {
	url := fmt.Sprintf("%s/bot%s/sendMessage", t.cfg.baseURL, t.cfg.botToken)

	payload := map[string]string{
		"chat_id":    t.cfg.chatID,
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

	resp, err := sharedClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram: unexpected status %d", resp.StatusCode)
	}
	return nil
}
