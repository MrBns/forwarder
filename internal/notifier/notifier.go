// Package notifier provides integrations that forward form submissions
// to external messaging services (Telegram, Discord).
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Notifier is the common interface implemented by every messaging back-end.
type Notifier interface {
	// Name returns a human-readable label for the notifier (e.g. "telegram").
	Name() string
	// Send delivers message text to the configured destination.
	Send(ctx context.Context, message string) error
}

// httpClient is the shared HTTP client used by all notifiers.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// -------------------------------------------------------------------
// Telegram
// -------------------------------------------------------------------

// TelegramNotifier forwards messages to a Telegram chat via Bot API.
type TelegramNotifier struct {
	botToken string
	chatID   string
	baseURL  string // overridable for tests; defaults to https://api.telegram.org
}

// NewTelegramNotifier creates a TelegramNotifier.
// Returns nil when either parameter is empty (disabled).
func NewTelegramNotifier(botToken, chatID string) *TelegramNotifier {
	return NewTelegramNotifierWithBaseURL(botToken, chatID, "https://api.telegram.org")
}

// NewTelegramNotifierWithBaseURL is like NewTelegramNotifier but lets callers
// override the Telegram API base URL (useful in tests).
func NewTelegramNotifierWithBaseURL(botToken, chatID, baseURL string) *TelegramNotifier {
	if botToken == "" || chatID == "" {
		return nil
	}
	return &TelegramNotifier{botToken: botToken, chatID: chatID, baseURL: baseURL}
}

func (t *TelegramNotifier) Name() string { return "telegram" }

func (t *TelegramNotifier) Send(ctx context.Context, message string) error {
	url := fmt.Sprintf("%s/bot%s/sendMessage", t.baseURL, t.botToken)

	payload := map[string]string{
		"chat_id":    t.chatID,
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

// -------------------------------------------------------------------
// Discord
// -------------------------------------------------------------------

// DiscordNotifier forwards messages to a Discord channel via webhook.
type DiscordNotifier struct {
	webhookURL string
}

// NewDiscordNotifier creates a DiscordNotifier.
// Returns nil when webhookURL is empty (disabled).
func NewDiscordNotifier(webhookURL string) *DiscordNotifier {
	if webhookURL == "" {
		return nil
	}
	return &DiscordNotifier{webhookURL: webhookURL}
}

func (d *DiscordNotifier) Name() string { return "discord" }

func (d *DiscordNotifier) Send(ctx context.Context, message string) error {
	payload := map[string]string{"content": message}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("discord: send request: %w", err)
	}
	defer resp.Body.Close()

	// Discord returns 204 No Content on success.
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord: unexpected status %d", resp.StatusCode)
	}
	return nil
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

// FormatFormData turns a map of form fields into a human-readable message.
func FormatFormData(origin string, fields map[string]string) string {
	var sb strings.Builder
	sb.WriteString("<b>📬 New Form Submission</b>\n")
	if origin != "" {
		sb.WriteString(fmt.Sprintf("<b>Origin:</b> %s\n", origin))
	}
	sb.WriteString("─────────────────\n")
	for k, v := range fields {
		sb.WriteString(fmt.Sprintf("<b>%s:</b> %s\n", k, v))
	}
	return sb.String()
}

// FormatFormDataPlain is the Discord-friendly (no HTML) variant.
func FormatFormDataPlain(origin string, fields map[string]string) string {
	var sb strings.Builder
	sb.WriteString("📬 **New Form Submission**\n")
	if origin != "" {
		sb.WriteString(fmt.Sprintf("**Origin:** %s\n", origin))
	}
	sb.WriteString("─────────────────\n")
	for k, v := range fields {
		sb.WriteString(fmt.Sprintf("**%s:** %s\n", k, v))
	}
	return sb.String()
}
