package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// TelegramNotifier sends messages to a Telegram chat via the Bot API.
type TelegramNotifier struct {
	botToken string
	chatID   string
	client   *http.Client
	// baseURL overrides the Telegram API host (used in tests).
	baseURL string
}

// NewTelegramNotifier creates a TelegramNotifier.  Returns nil when either
// botToken or chatID is empty so the caller can skip registration.
func NewTelegramNotifier(botToken, chatID string, client *http.Client) *TelegramNotifier {
	if botToken == "" || chatID == "" {
		return nil
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatID,
		client:   client,
		baseURL:  "https://api.telegram.org",
	}
}

// NewTelegramNotifierWithBaseURL creates a TelegramNotifier with a custom base
// URL, intended for use in tests.
func NewTelegramNotifierWithBaseURL(botToken, chatID string, client *http.Client, baseURL string) *TelegramNotifier {
	n := NewTelegramNotifier(botToken, chatID, client)
	if n == nil {
		return nil
	}
	n.baseURL = baseURL
	return n
}

func (t *TelegramNotifier) Name() string { return "telegram" }

// Send formats the Message as an HTML Telegram message and POSTs it.
func (t *TelegramNotifier) Send(msg Message) error {
	text := buildTelegramText(msg)

	payload := map[string]string{
		"chat_id":    t.chatID,
		"text":       text,
		"parse_mode": "HTML",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", t.baseURL, t.botToken)
	resp, err := t.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram: unexpected status %d", resp.StatusCode)
	}
	return nil
}

// buildTelegramText converts a Message into an HTML-formatted Telegram string.
func buildTelegramText(msg Message) string {
	var sb strings.Builder

	if msg.Title != "" {
		sb.WriteString("<b>")
		sb.WriteString(htmlEscape(msg.Title))
		sb.WriteString("</b>\n\n")
	}

	if msg.Description != "" {
		sb.WriteString(htmlEscape(msg.Description))
		sb.WriteString("\n\n")
	}

	if msg.Note != "" {
		sb.WriteString("📝 <i>Note: ")
		sb.WriteString(htmlEscape(msg.Note))
		sb.WriteString("</i>\n\n")
	}

	if len(msg.Fields) > 0 {
		sb.WriteString("<b>Details:</b>\n")
		for k, v := range msg.Fields {
			sb.WriteString(fmt.Sprintf("• <b>%s:</b> %s\n", htmlEscape(k), htmlEscape(v)))
		}
		sb.WriteString("\n")
	}

	if len(msg.Attachments) > 0 {
		sb.WriteString("<b>Attachments:</b>\n")
		for _, a := range msg.Attachments {
			name := a.Name
			if name == "" {
				name = a.URL
			}
			sb.WriteString(fmt.Sprintf("• <a href=\"%s\">%s</a>\n", a.URL, htmlEscape(name)))
		}
		sb.WriteString("\n")
	}

	if msg.Footer != "" {
		sb.WriteString("—\n")
		sb.WriteString("<i>")
		sb.WriteString(htmlEscape(msg.Footer))
		sb.WriteString("</i>")
	}

	return strings.TrimSpace(sb.String())
}

// htmlEscape escapes characters that have special meaning in Telegram HTML mode.
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
