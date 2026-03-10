package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// SlackNotifier sends messages to a Slack channel via an incoming webhook.
type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewSlackNotifier creates a SlackNotifier.  Returns nil when the webhook URL
// is empty so the caller can skip registration.
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
	Type string    `json:"type"`
	Text *slackText `json:"text,omitempty"`
}

type slackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Send formats the Message as a Slack Block Kit payload and POSTs it.
func (s *SlackNotifier) Send(msg Message) error {
	blocks := buildSlackBlocks(msg)

	payload := map[string]interface{}{"blocks": blocks}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}

// buildSlackBlocks converts a Message into Slack Block Kit blocks.
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
