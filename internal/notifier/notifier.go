package notifier

// Attachment represents a linked file or URL to include with the forwarded message.
type Attachment struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	// Type hints the kind of attachment: "image", "file", "link", etc.
	Type string `json:"type,omitempty"`
}

// Message is the unified, platform-agnostic payload that the forwarder sends
// to each enabled notifier.
type Message struct {
	// Title is a short heading for the notification.
	Title string `json:"title,omitempty"`

	// Description is the main body text of the notification.
	Description string `json:"description,omitempty"`

	// Note is an optional internal remark or highlight (shown as a quote or callout).
	Note string `json:"note,omitempty"`

	// Footer is a small piece of metadata shown at the bottom of the notification.
	Footer string `json:"footer,omitempty"`

	// Attachments is a list of files or links appended to the notification.
	Attachments []Attachment `json:"attachments,omitempty"`

	// Fields is an arbitrary key-value map for extra structured data.
	Fields map[string]string `json:"fields,omitempty"`
}

// Result captures whether a specific notifier succeeded and the relevant detail.
type Result struct {
	Platform string `json:"platform"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

// Notifier is the interface every messaging-platform adapter must implement.
type Notifier interface {
	// Name returns the human-readable platform name (e.g. "discord").
	Name() string

	// Send delivers msg to the platform.
	Send(msg Message) error
}
