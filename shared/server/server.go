// Package server provides the shared HTTP router factory and the health-check
// handler.  Feature routes are registered by the composition root after
// calling NewRouter so the server package stays free of feature imports.
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/MrBns/forwarder/shared/config"
)

// NewRouter creates a chi.Mux pre-loaded with the standard middleware stack
// and CORS policy derived from cfg.  The composition root registers all
// feature routes on the returned mux before handing it to http.Server.
func NewRouter(cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	return r
}

// HealthCheck handles GET /health — a simple liveness probe.
func HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
