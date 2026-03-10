package main

// container.go is the composition root: it wires all feature dependencies
// together using explicit constructor injection — no code generation, no
// reflection framework.  Each dependency is constructed exactly once and
// passed to every consumer that needs it.
//
// Dependency graph
//
//	config.Load()
//	  ├─ forwarder feature
//	  │    ├─ NewDiscordNotifier  (nil when DISCORD_WEBHOOK_URL is unset)
//	  │    ├─ NewSlackNotifier    (nil when SLACK_WEBHOOK_URL is unset)
//	  │    ├─ NewTelegramNotifier (nil when token/chatID are unset)
//	  │    └─ NewHandler          → POST /api/forward
//	  ├─ submitter feature
//	  │    ├─ NewTelegramNotifier (nil when token/chatID are unset)
//	  │    ├─ NewDiscordNotifier  (nil when DISCORD_WEBHOOK_URL is unset)
//	  │    ├─ NewNotifiers
//	  │    └─ NewHandler          → POST /api/submit
//	  └─ feedback feature  (only when DATABASE_URL is set)
//	       ├─ db.NewPool
//	       ├─ NewRepository
//	       ├─ NewService
//	       └─ NewHandler         → POST /api/feedback, GET /api/feedback
//
//	server.NewRouter(cfg)          ← receives AllowedOrigins for CORS
//	  ├─ GET  /health
//	  ├─ POST /api/forward
//	  ├─ POST /api/submit
//	  ├─ POST /api/feedback        (registered only when DATABASE_URL is set)
//	  └─ GET  /api/feedback        (registered only when DATABASE_URL is set)

import (
	"log"
	"net/http"

	"github.com/mrbns/forwarder/internal/feature/feedback"
	"github.com/mrbns/forwarder/internal/feature/forwarder"
	"github.com/mrbns/forwarder/internal/feature/submitter"
	"github.com/mrbns/forwarder/shared/config"
	"github.com/mrbns/forwarder/shared/db"
	"github.com/mrbns/forwarder/shared/server"
)

// build resolves the full dependency graph and returns a configured http.Handler.
func build(cfg *config.Config) (http.Handler, error) {
	// ── forwarder feature ─────────────────────────────────────────────────────
	// Outgoing adapters — nil when credentials are absent (adapter auto-disabled).
	var fwdNotifiers []forwarder.Notifier
	if d := forwarder.NewDiscordNotifier(cfg.DiscordWebhookURL, nil); d != nil {
		fwdNotifiers = append(fwdNotifiers, d)
	}
	if s := forwarder.NewSlackNotifier(cfg.SlackWebhookURL, nil); s != nil {
		fwdNotifiers = append(fwdNotifiers, s)
	}
	if t := forwarder.NewTelegramNotifier(cfg.TelegramBotToken, cfg.TelegramChatID, nil); t != nil {
		fwdNotifiers = append(fwdNotifiers, t)
	}
	// Incoming adapter
	forwardHandler := forwarder.NewHandler(fwdNotifiers)

	// ── submitter feature ─────────────────────────────────────────────────────
	// Outgoing adapters
	tgNotifier := submitter.NewTelegramNotifier(cfg.TelegramBotToken, cfg.TelegramChatID)
	dcNotifier := submitter.NewDiscordNotifier(cfg.DiscordWebhookURL)
	submitNotifiers := submitter.NewNotifiers(tgNotifier, dcNotifier)
	// Incoming adapter
	submitHandler := submitter.NewHandler(submitNotifiers)

	// ── router ───────────────────────────────────────────────────────────────
	r := server.NewRouter(cfg)
	r.Get("/health", server.HealthCheck)
	r.Post("/api/forward", forwardHandler.ServeHTTP)
	r.Post("/api/submit", submitHandler.Submit)

	// ── feedback feature  (optional — requires DATABASE_URL) ─────────────────
	if cfg.DatabaseURL != "" {
		pool, err := db.NewPool(cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
		repo, err := feedback.NewRepository(pool)
		if err != nil {
			return nil, err
		}
		svc := feedback.NewService(repo)
		fbHandler := feedback.NewHandler(svc)

		r.Post("/api/feedback", fbHandler.Submit)
		r.Get("/api/feedback", fbHandler.List)

		log.Println("feedback: enabled (PostgreSQL)")
	} else {
		log.Println("feedback: disabled (DATABASE_URL not set)")
	}

	return r, nil
}
