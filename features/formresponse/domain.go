// Package formresponse implements the form-response feature.
// It follows a Hexagonal Architecture layout:
//
//   - domain.go      — outgoing port (Notifier interface) and domain types
//   - handler.go     — incoming adapter (HTTP → domain)
//   - telegram.go    — outgoing adapter (domain → Telegram Bot API)
//   - discord.go     — outgoing adapter (domain → Discord webhook)
//   - providers.go   — constructor helpers for the Notifiers slice
package formresponse

import (
	"context"
	"fmt"
	"strings"
)

// Notifier is the outgoing port for delivering form-submission notifications.
// Any messaging back-end (Telegram, Discord, …) must implement this interface.
type Notifier interface {
	// Name returns a human-readable label used in logs (e.g. "telegram").
	Name() string
	// Send delivers the formatted message to the destination.
	Send(ctx context.Context, message string) error
}

// Notifiers is an ordered slice of active Notifier instances.
// It is a distinct named type so Wire can inject it unambiguously.
type Notifiers []Notifier

// FormatHTML renders form fields as an HTML-formatted string (Telegram).
func FormatHTML(origin string, fields map[string]string) string {
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

// FormatMarkdown renders form fields as a Markdown-formatted string (Discord).
func FormatMarkdown(origin string, fields map[string]string) string {
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
