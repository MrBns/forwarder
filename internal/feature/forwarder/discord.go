package forwarder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DiscordNotifier is the outgoing adapter that delivers messages to a Discord
// channel via an incoming webhook using Markdown formatting.
// It implements the Notifier port.
type DiscordNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewDiscordNotifier creates a DiscordNotifier.
// Returns nil when webhookURL is empty (adapter disabled).
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

// Send formats msg as Markdown and POSTs it to the Discord webhook.
func (d *DiscordNotifier) Send(msg Message) ([]byte, error) {
	content := buildDiscordContent(msg)

	payload := map[string]string{"content": content}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("discord: marshal payload: %w", err)
	}

	resp, err := d.client.Post(d.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("discord: send request: %w", err)
	}

	data, err := io.ReadAll(resp.Body)

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return data, fmt.Errorf("discord: unexpected status %d", resp.StatusCode)
	}
	return data, err
}

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
		sb.WriteString("---\n*")
		sb.WriteString(msg.Footer)
		sb.WriteString("*")
	}

	return strings.TrimSpace(sb.String())
}
