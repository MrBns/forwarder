package feedback

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// Handler is the incoming HTTP adapter for the feedback feature.
// It translates HTTP requests into calls on the Service incoming port.
type Handler struct {
	svc Service
}

// NewHandler creates a Handler backed by the supplied Service.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// submitRequest is the JSON body accepted by POST /api/feedback.
type submitRequest struct {
	Fields map[string]string `json:"fields"`
}

// Submit handles POST /api/feedback — validates the request, calls the
// service, and returns the persisted Feedback as JSON.
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResponse("invalid JSON body"))
		return
	}

	if len(req.Fields) == 0 {
		writeJSON(w, http.StatusBadRequest, errResponse("fields must not be empty"))
		return
	}

	origin := r.Header.Get("Origin")
	fb, err := h.svc.Submit(r.Context(), origin, req.Fields)
	if err != nil {
		log.Printf("feedback submit error: %v", err)
		writeJSON(w, http.StatusInternalServerError, errResponse("failed to save feedback"))
		return
	}

	writeJSON(w, http.StatusCreated, fb)
}

// List handles GET /api/feedback — returns paginated feedbacks.
// Query params: limit (default 20), offset (default 0).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	fbs, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		log.Printf("feedback list error: %v", err)
		writeJSON(w, http.StatusInternalServerError, errResponse("failed to retrieve feedbacks"))
		return
	}

	if fbs == nil {
		fbs = []*Feedback{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"feedbacks": fbs})
}

// queryInt reads an integer query parameter, returning defaultVal on absence
// or parse error.
func queryInt(r *http.Request, key string, defaultVal int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}

func errResponse(msg string) map[string]any {
	return map[string]any{"success": false, "message": msg}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}
