package feedback

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// Handler is the incoming HTTP adapter for the feedback feature.
type Handler struct {
	svc Service
}

// NewHandler creates a Handler backed by the supplied Service.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

type submitRequest struct {
	Fields map[string]string `json:"fields"`
}

// Submit handles POST /api/feedback.
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	if len(req.Fields) == 0 {
		writeJSON(w, http.StatusBadRequest, errBody("fields must not be empty"))
		return
	}

	origin := r.Header.Get("Origin")
	fb, err := h.svc.Submit(r.Context(), origin, req.Fields)
	if err != nil {
		log.Printf("feedback submit error: %v", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to save feedback"))
		return
	}
	writeJSON(w, http.StatusCreated, fb)
}

// List handles GET /api/feedback with optional ?limit=N&offset=M query params.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	fbs, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		log.Printf("feedback list error: %v", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to retrieve feedbacks"))
		return
	}
	if fbs == nil {
		fbs = []*Feedback{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"feedbacks": fbs})
}

func queryInt(r *http.Request, key string, def int) int {
	if raw := r.URL.Query().Get(key); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v >= 0 {
			return v
		}
	}
	return def
}

func errBody(msg string) map[string]any {
	return map[string]any{"success": false, "message": msg}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}
