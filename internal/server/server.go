package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/mrbns/forwarder/internal/config"
	"github.com/mrbns/forwarder/internal/handler"
	"github.com/mrbns/forwarder/internal/notifier"
)

// New assembles and returns the application HTTP router.
func New(cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Build enabled notifiers from configuration.
	notifiers := buildNotifiers(cfg)

	forwardHandler := handler.NewForwardHandler(notifiers)

	// Routes
	r.Get("/health", healthHandler)
	r.Post("/api/forward", forwardHandler.ServeHTTP)

	return r
}

// buildNotifiers constructs only the notifiers whose credentials are present.
func buildNotifiers(cfg *config.Config) []notifier.Notifier {
	var notifiers []notifier.Notifier

	if d := notifier.NewDiscordNotifier(cfg.DiscordWebhookURL, nil); d != nil {
		notifiers = append(notifiers, d)
	}
	if s := notifier.NewSlackNotifier(cfg.SlackWebhookURL, nil); s != nil {
		notifiers = append(notifiers, s)
	}
	if t := notifier.NewTelegramNotifier(cfg.TelegramBotToken, cfg.TelegramChatID, nil); t != nil {
		notifiers = append(notifiers, t)
	}

	return notifiers
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
