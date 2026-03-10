// Package submitter implements the form-submission forwarding feature.
//
// Hexagonal Architecture layout:
//
//   - domain.go    — Notifier outgoing port and message-formatting helpers
//   - handler.go   — HTTP incoming adapter  (POST /api/submit)
//   - telegram.go  — Telegram outgoing adapter
//   - discord.go   — Discord outgoing adapter
//   - notifiers.go — NewNotifiers constructor helper
package submitter

import (
	"context"
	"fmt"
	"strings"
)

// Notifier is the outgoing port for delivering form-submission notifications.
// Any messaging back-end must implement this interface to participate.
type Notifier interface {
	// Name returns a human-readable label used in logs (e.g. "telegram").
	Name() string
	// Send delivers the formatted message string to the destination.
	Send(ctx context.Context, message string) error
}

// Notifiers is a named slice of active Notifier instances.
type Notifiers []Notifier

// FormatHTML renders form fields as an HTML string for Telegram.
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

// FormatMarkdown renders form fields as a Markdown string for Discord.
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
