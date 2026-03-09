package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// DiscordNotifier sends messages to a Discord channel via an incoming webhook.
type DiscordNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewDiscordNotifier creates a DiscordNotifier.  Returns nil when the webhook
// URL is empty so the caller can skip registration.
func NewDiscordNotifier(webhookURL string, client *http.Client) *DiscordNotifier {
	if webhookURL == "" {
		return nil
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &DiscordNotifier{webhookURL: webhookURL, client: client}
}

func (d *DiscordNotifier) Name() string { return "discord" }

// Send formats the Message as a Discord webhook payload and POSTs it.
func (d *DiscordNotifier) Send(msg Message) error {
	content := buildDiscordContent(msg)

	payload := map[string]string{"content": content}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord: marshal payload: %w", err)
	}

	resp, err := d.client.Post(d.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord: unexpected status %d", resp.StatusCode)
	}
	return nil
}

// buildDiscordContent converts a Message into a Markdown-formatted string.
func buildDiscordContent(msg Message) string {
	var sb strings.Builder

	if msg.Title != "" {
		sb.WriteString("## ")
		sb.WriteString(msg.Title)
		sb.WriteString("\n\n")
	}

	if msg.Description != "" {
		sb.WriteString(msg.Description)
		sb.WriteString("\n\n")
	}

	if msg.Note != "" {
		sb.WriteString("> **Note:** ")
		sb.WriteString(msg.Note)
		sb.WriteString("\n\n")
	}

	if len(msg.Fields) > 0 {
		sb.WriteString("**Details:**\n")
		for k, v := range msg.Fields {
			sb.WriteString(fmt.Sprintf("• **%s:** %s\n", k, v))
		}
		sb.WriteString("\n")
	}

	if len(msg.Attachments) > 0 {
		sb.WriteString("**Attachments:**\n")
		for _, a := range msg.Attachments {
			name := a.Name
			if name == "" {
				name = a.URL
			}
			sb.WriteString(fmt.Sprintf("• [%s](%s)\n", name, a.URL))
		}
		sb.WriteString("\n")
	}

	if msg.Footer != "" {
		sb.WriteString("---\n")
		sb.WriteString("*")
		sb.WriteString(msg.Footer)
		sb.WriteString("*")
	}

	return strings.TrimSpace(sb.String())
}
