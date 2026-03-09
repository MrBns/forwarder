package formresponse

import (
	"encoding/json"
	"log"
	"net/http"
)

// FormHandler is the incoming HTTP adapter for the formresponse feature.
// It receives form submissions over HTTP and dispatches them to all active
// Notifier outgoing adapters.
type FormHandler struct {
	notifiers Notifiers
}

// NewFormHandler creates a FormHandler with the injected Notifiers slice.
func NewFormHandler(notifiers Notifiers) *FormHandler {
	return &FormHandler{notifiers: notifiers}
}

// submitRequest is the JSON body accepted by the submit endpoint.
type submitRequest struct {
	// Fields contains arbitrary key-value pairs from the HTML form.
	Fields map[string]string `json:"fields"`
}

// submitResponse is the JSON body returned on success or failure.
type submitResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Submit handles POST /api/submit — accepts a form payload and forwards it
// to every configured notifier.
func (h *FormHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, submitResponse{
			Success: false,
			Message: "invalid JSON body",
		})
		return
	}

	if len(req.Fields) == 0 {
		writeJSON(w, http.StatusBadRequest, submitResponse{
			Success: false,
			Message: "fields must not be empty",
		})
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

	writeJSON(w, http.StatusOK, submitResponse{
		Success: true,
		Message: "form submitted successfully",
	})
}

// HealthCheck handles GET /health — returns a simple liveness probe.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}
