package submitter

import (
	"encoding/json"
	"log"
	"net/http"
)

// Handler is the incoming HTTP adapter for the submitter feature.
// It decodes form payloads and dispatches them to all active Notifiers.
type Handler struct {
	notifiers Notifiers
}

// NewHandler creates a Handler backed by the supplied Notifiers slice.
func NewHandler(notifiers Notifiers) *Handler {
	return &Handler{notifiers: notifiers}
}

type submitRequest struct {
	Fields map[string]string `json:"fields"`
}

type submitResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Submit handles POST /api/submit — accepts a form payload and forwards it to
// every configured Notifier.
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, submitResponse{Success: false, Message: "invalid JSON body"})
		return
	}
	if len(req.Fields) == 0 {
		writeJSON(w, http.StatusBadRequest, submitResponse{Success: false, Message: "fields must not be empty"})
		return
	}

	origin := r.Header.Get("Origin")
	ctx := r.Context()

	var failed []string
	for _, n := range h.notifiers {
		var msg string
		if n.Name() == "discord" {
			msg = FormatMarkdown(origin, req.Fields)
		} else {
			msg = FormatHTML(origin, req.Fields)
		}
		if err := n.Send(ctx, msg); err != nil {
			log.Printf("notifier %s error: %v", n.Name(), err)
			failed = append(failed, n.Name())
		}
	}

	if len(failed) > 0 && len(failed) == len(h.notifiers) {
		writeJSON(w, http.StatusInternalServerError, submitResponse{
			Success: false,
			Message: "failed to deliver notification",
		})
		return
	}

	writeJSON(w, http.StatusOK, submitResponse{Success: true, Message: "form submitted successfully"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}
