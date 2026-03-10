package forwarder

import (
	"encoding/json"
	"net/http"
)

// Request is the JSON body accepted by POST /api/forward.
type Request struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Note        string            `json:"note"`
	Footer      string            `json:"footer"`
	Attachments []Attachment      `json:"attachments"`
	Fields      map[string]string `json:"fields"`
	// Platforms is an optional list of platform names to target.
	// When omitted or empty all enabled notifiers receive the message.
	Platforms []string `json:"platforms"`
}

// Response is the JSON body returned by POST /api/forward.
type Response struct {
	Results []Result `json:"results"`
}

// Handler is the incoming HTTP adapter for the forwarder feature.
// It decodes the request, resolves target notifiers, fans the message out,
// and returns per-platform delivery results.
type Handler struct {
	notifiers []Notifier
}

// NewHandler creates a Handler backed by the supplied notifier slice.
func NewHandler(notifiers []Notifier) *Handler {
	return &Handler{notifiers: notifiers}
}

// ServeHTTP handles POST /api/forward.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON body: " + err.Error(),
		})
		return
	}

	msg := Message{
		Title:       req.Title,
		Description: req.Description,
		Note:        req.Note,
		Footer:      req.Footer,
		Attachments: req.Attachments,
		Fields:      req.Fields,
	}

	targets := h.resolveTargets(req.Platforms)
	if len(targets) == 0 {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "no platforms are enabled or matched the requested platforms",
		})
		return
	}

	results := make([]Result, 0, len(targets))
	for _, n := range targets {
		res := Result{Platform: n.Name(), Success: true}
		if err := n.Send(msg); err != nil {
			res.Success = false
			res.Error = err.Error()
		}
		results = append(results, res)
	}

	writeJSON(w, http.StatusOK, Response{Results: results})
}

// resolveTargets returns the notifiers that match the requested platform list.
// When platforms is nil/empty, all registered notifiers are returned.
func (h *Handler) resolveTargets(platforms []string) []Notifier {
	if len(platforms) == 0 {
		return h.notifiers
	}
	set := make(map[string]struct{}, len(platforms))
	for _, p := range platforms {
		set[p] = struct{}{}
	}
	var targets []Notifier
	for _, n := range h.notifiers {
		if _, ok := set[n.Name()]; ok {
			targets = append(targets, n)
		}
	}
	return targets
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
