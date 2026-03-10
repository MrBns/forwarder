// Package forwarder implements the unified notification-forwarding feature.
//
// Hexagonal Architecture layout:
//
//   - domain.go   — domain types (Message, Attachment, Result) and the
//     Notifier outgoing port
//   - handler.go  — HTTP incoming adapter  (POST /api/forward)
//   - telegram.go — Telegram outgoing adapter
//   - discord.go  — Discord outgoing adapter
//   - slack.go    — Slack outgoing adapter
package forwarder

import "encoding/json"

// Attachment represents a linked file or URL included with a forwarded message.
type Attachment struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	// Type hints the kind of attachment: "image", "file", "link", etc.
	Type string `json:"type,omitempty"`
}

// Message is the unified, platform-agnostic notification payload that the
// handler builds and every outgoing adapter receives.
type Message struct {
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	Note        string            `json:"note,omitempty"`
	Footer      string            `json:"footer,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Fields      map[string]string `json:"fields,omitempty"`
}

// Result captures the delivery outcome for a single platform.
type Result struct {
	Platform string          `json:"platform"`
	Success  bool            `json:"success"`
	Error    string          `json:"error,omitempty"`
	Data     json.RawMessage `json:"data"`
}

// Notifier is the outgoing port that every messaging-platform adapter must
// implement.  New platforms are added by creating an adapter that satisfies
// this interface and registering it in the composition root.
type Notifier interface {
	// Name returns the human-readable platform identifier (e.g. "discord").
	Name() string
	// Send delivers msg to the platform.  It returns an error when delivery
	// fails; partial failures are surfaced in the API response, not panicked.
	Send(msg Message) ([]byte, error)
}
