package forwarder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// TelegramNotifier is the outgoing adapter that delivers messages via the
// Telegram Bot API using HTML formatting.  It implements the Notifier port.
type TelegramNotifier struct {
	botToken string
	chatID   string
	client   *http.Client
	baseURL  string
}

// NewTelegramNotifier creates a production TelegramNotifier.
// Returns nil when botToken or chatID is empty (adapter disabled).
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

// NewTelegramNotifierWithBaseURL creates a TelegramNotifier with a custom API
// base URL, intended for use in tests that spin up a fake Telegram server.
func NewTelegramNotifierWithBaseURL(botToken, chatID string, client *http.Client, baseURL string) *TelegramNotifier {
	n := NewTelegramNotifier(botToken, chatID, client)
	if n == nil {
		return nil
	}
	n.baseURL = baseURL
	return n
}

func (t *TelegramNotifier) Name() string { return "telegram" }

// Send formats msg as HTML and POSTs it to the Telegram Bot API.
func (t *TelegramNotifier) Send(msg Message) ([]byte, error) {
	text := buildTelegramText(msg)

	payload := map[string]string{
		"chat_id":    t.chatID,
		"text":       text,
		"parse_mode": "HTML",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("telegram: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", t.baseURL, t.botToken)
	resp, err := t.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("telegram: send request: %w", err)
	}

	data, err := io.ReadAll(resp.Body)

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return data, fmt.Errorf("telegram: unexpected status %d", resp.StatusCode)
	}

	return data, err
}

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
		sb.WriteString("<i>📝 Note: ")
		sb.WriteString(htmlEscape(msg.Note))
		sb.WriteString("</i>\n\n")
	}
	if len(msg.Fields) > 0 {
		sb.WriteString("<b>ℹ️ Details:</b>\n")
		for k, v := range msg.Fields {
			sb.WriteString(fmt.Sprintf("• <b>%s:</b> %s\n", htmlEscape(k), htmlEscape(v)))
		}
		sb.WriteString("\n")
	}
	if len(msg.Attachments) > 0 {
		sb.WriteString("<b>🔗 Attachments:</b>\n")
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
		sb.WriteString("<i>")
		sb.WriteString(htmlEscape(msg.Footer))
		sb.WriteString("</i>")
	}

	return strings.TrimSpace(sb.String())
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
