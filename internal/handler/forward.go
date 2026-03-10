package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mrbns/forwarder/internal/notifier"
)

// ForwardRequest is the JSON body accepted by POST /api/forward.
type ForwardRequest struct {
	// Title is a short heading for the notification.
	Title string `json:"title"`

	// Description is the main body text.
	Description string `json:"description"`

	// Note is an optional internal remark shown as a callout.
	Note string `json:"note"`

	// Footer is small metadata shown at the bottom of the notification.
	Footer string `json:"footer"`

	// Attachments is a list of files or links appended to the notification.
	Attachments []notifier.Attachment `json:"attachments"`

	// Fields is an arbitrary key-value map for extra structured data.
	Fields map[string]string `json:"fields"`

	// Platforms is an optional list of platform names to target.
	// When omitted or empty all enabled platforms are used.
	Platforms []string `json:"platforms"`
}

// ForwardResponse is the JSON body returned by POST /api/forward.
type ForwardResponse struct {
	Results []notifier.Result `json:"results"`
}

// ForwardHandler handles POST /api/forward.
type ForwardHandler struct {
	notifiers []notifier.Notifier
}

// NewForwardHandler creates a ForwardHandler with the provided notifiers.
func NewForwardHandler(notifiers []notifier.Notifier) *ForwardHandler {
	return &ForwardHandler{notifiers: notifiers}
}

// ServeHTTP implements http.Handler.
func (h *ForwardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req ForwardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON body: " + err.Error(),
		})
		return
	}

	msg := notifier.Message{
		Title:       req.Title,
		Description: req.Description,
		Note:        req.Note,
		Footer:      req.Footer,
		Attachments: req.Attachments,
		Fields:      req.Fields,
	}

	// Determine which notifiers to invoke.
	targets := h.resolveTargets(req.Platforms)
	if len(targets) == 0 {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "no platforms are enabled or matched the requested platforms",
		})
		return
	}

	results := make([]notifier.Result, 0, len(targets))
	for _, n := range targets {
		r := notifier.Result{Platform: n.Name(), Success: true}
		if err := n.Send(msg); err != nil {
			r.Success = false
			r.Error = err.Error()
		}
		results = append(results, r)
	}

	writeJSON(w, http.StatusOK, ForwardResponse{Results: results})
}

// resolveTargets returns the notifiers that match the requested platform list.
// When platforms is nil/empty, all enabled notifiers are returned.
func (h *ForwardHandler) resolveTargets(platforms []string) []notifier.Notifier {
	if len(platforms) == 0 {
		return h.notifiers
	}

	set := make(map[string]struct{}, len(platforms))
	for _, p := range platforms {
		set[p] = struct{}{}
	}

	var targets []notifier.Notifier
	for _, n := range h.notifiers {
		if _, ok := set[n.Name()]; ok {
			targets = append(targets, n)
		}
	}
	return targets
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
