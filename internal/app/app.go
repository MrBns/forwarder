// Package app wires together the HTTP router and all middleware to produce the
// top-level application that main.go starts.
package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/wire"

	"github.com/MrBns/form-response/internal/config"
	"github.com/MrBns/form-response/internal/handler"
	"github.com/MrBns/form-response/internal/notifier"
)

// ProviderSet is the Wire provider set for the app package.
var ProviderSet = wire.NewSet(New)

// App is the fully initialised application ready to serve HTTP traffic.
type App struct {
	router http.Handler
	port   string
}

// New constructs the App by wiring the router, middleware, and routes.
// Wire injects the ServerConfig, FormHandler, and Notifiers at compile time.
func New(cfg config.ServerConfig, formHandler *handler.FormHandler, ns notifier.Notifiers) *App {
	r := chi.NewRouter()

	// Standard middleware stack.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS — restrict to configured origins.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Routes.
	r.Get("/health", handler.HealthCheck)
	r.Route("/api", func(r chi.Router) {
		r.Post("/submit", formHandler.Submit)
	})

	// Log which notifiers are active.
	if len(ns) == 0 {
		log.Println("WARNING: no notifier configured – set TELEGRAM_BOT_TOKEN/TELEGRAM_CHAT_ID or DISCORD_WEBHOOK_URL")
	}
	for _, n := range ns {
		log.Printf("%s notifier: enabled", n.Name())
	}

	return &App{router: r, port: cfg.Port}
}

// Run starts the HTTP server and blocks until it exits.
func (a *App) Run() error {
	addr := fmt.Sprintf(":%s", a.port)
	log.Printf("form-response API listening on %s", addr)
	return http.ListenAndServe(addr, a.router)
}
