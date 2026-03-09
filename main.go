package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/MrBns/form-response/internal/config"
	"github.com/MrBns/form-response/internal/handler"
	"github.com/MrBns/form-response/internal/notifier"
)

func main() {
	cfg := config.Load()

	// Build notifiers (nil ones are silently skipped inside NewFormHandler).
	tgNotifier := notifier.NewTelegramNotifier(cfg.TelegramBotToken, cfg.TelegramChatID)
	dcNotifier := notifier.NewDiscordNotifier(cfg.DiscordWebhookURL)

	if tgNotifier == nil && dcNotifier == nil {
		log.Println("WARNING: no notifier configured – set TELEGRAM_BOT_TOKEN/TELEGRAM_CHAT_ID or DISCORD_WEBHOOK_URL")
	}

	formHandler := handler.NewFormHandler(tgNotifier, dcNotifier)

	r := chi.NewRouter()

	// Standard middleware stack.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS – restrict to configured origins.
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

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("form-response API listening on %s", addr)
	log.Printf("allowed origins: %v", cfg.AllowedOrigins)
	if tgNotifier != nil {
		log.Println("Telegram notifier: enabled")
	}
	if dcNotifier != nil {
		log.Println("Discord notifier:  enabled")
	}

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
