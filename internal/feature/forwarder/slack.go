package forwarder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SlackNotifier is the outgoing adapter that delivers messages to a Slack
// channel via an incoming webhook using Block Kit formatting.
// It implements the Notifier port.
type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewSlackNotifier creates a SlackNotifier.
// Returns nil when webhookURL is empty (adapter disabled).
func NewSlackNotifier(webhookURL string, client *http.Client) *SlackNotifier {
	if webhookURL == "" {
		return nil
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &SlackNotifier{webhookURL: webhookURL, client: client}
}

func (s *SlackNotifier) Name() string { return "slack" }

// slackBlock is a simplified Slack Block Kit block.
type slackBlock struct {
	Type string     `json:"type"`
	Text *slackText `json:"text,omitempty"`
}

type slackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Send formats msg as Slack Block Kit blocks and POSTs it to the webhook.
func (s *SlackNotifier) Send(msg Message) ([]byte, error) {
	blocks := buildSlackBlocks(msg)

	payload := map[string]any{"blocks": blocks}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("slack: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("slack: send request: %w", err)
	}

	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return data, fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return data, err
}

func buildSlackBlocks(msg Message) []slackBlock {
	var blocks []slackBlock

	if msg.Title != "" {
		blocks = append(blocks, slackBlock{
			Type: "header",
			Text: &slackText{Type: "plain_text", Text: msg.Title},
		})
	}
	if msg.Description != "" {
		blocks = append(blocks, slackBlock{
			Type: "section",
			Text: &slackText{Type: "mrkdwn", Text: msg.Description},
		})
	}
	if msg.Note != "" {
		blocks = append(blocks, slackBlock{
			Type: "section",
			Text: &slackText{Type: "mrkdwn", Text: "> :memo: *Note:* " + msg.Note},
		})
	}
	if len(msg.Fields) > 0 {
		var sb strings.Builder
		sb.WriteString("*Details:*\n")
		for k, v := range msg.Fields {
			sb.WriteString(fmt.Sprintf("• *%s:* %s\n", k, v))
		}
		blocks = append(blocks, slackBlock{
			Type: "section",
			Text: &slackText{Type: "mrkdwn", Text: strings.TrimRight(sb.String(), "\n")},
		})
	}
	if len(msg.Attachments) > 0 {
		var sb strings.Builder
		sb.WriteString("*Attachments:*\n")
		for _, a := range msg.Attachments {
			name := a.Name
			if name == "" {
				name = a.URL
			}
			sb.WriteString(fmt.Sprintf("• <%s|%s>\n", a.URL, name))
		}
		blocks = append(blocks, slackBlock{
			Type: "section",
			Text: &slackText{Type: "mrkdwn", Text: strings.TrimRight(sb.String(), "\n")},
		})
	}
	if msg.Footer != "" {
		blocks = append(blocks, slackBlock{
			Type: "context",
			Text: &slackText{Type: "mrkdwn", Text: msg.Footer},
		})
	}

	return blocks
}
