package formresponse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// DiscordNotifier is the outgoing adapter that forwards messages via a
// Discord incoming webhook.  It implements the Notifier outgoing port.
type DiscordNotifier struct {
	webhookURL string
}

// NewDiscordNotifier creates a DiscordNotifier.
// Returns nil when webhookURL is empty (feature disabled).
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
