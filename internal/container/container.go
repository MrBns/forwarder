// Package container is the composition root of the application.
//
// It wires every dependency together using plain constructor injection — no
// code generation, no reflection framework.  Each dependency is constructed
// once and passed explicitly to its consumers, following the Dependency
// Inversion principle at every layer boundary.
//
// Dependency graph (top-level → leaf):
//
//	config.Load()
//	  ├── ServerConfig        → app.New
//	  ├── TelegramConfig      → formresponse.NewTelegramNotifier
//	  ├── DiscordConfig       → formresponse.NewDiscordNotifier
//	  └── DBConfig            → db.NewPool
//	                                └── feedback.NewRepository
//	                                      └── feedback.NewService
//	                                            └── feedback.NewHandler → app.New
//	formresponse.NewNotifiers → formresponse.NewFormHandler → app.New
package container

import (
	"github.com/MrBns/forwarder/features/feedback"
	"github.com/MrBns/forwarder/features/formresponse"
	"github.com/MrBns/forwarder/internal/app"
	"github.com/MrBns/forwarder/internal/config"
	"github.com/MrBns/forwarder/internal/db"
)

// Build resolves the full dependency graph and returns a configured *app.App.
// Each dependency is constructed exactly once and injected into its consumers.
func Build() (*app.App, error) {
	// ── Configuration ────────────────────────────────────────────────────────
	cfg := config.Load()
	serverCfg := config.ProvideServerConfig(cfg)
	telegramCfg := config.ProvideTelegramConfig(cfg)
	discordCfg := config.ProvideDiscordConfig(cfg)
	dbCfg := config.ProvideDBConfig(cfg)

	// ── formresponse feature ──────────────────────────────────────────────────
	// Outgoing adapters (implement the Notifier outgoing port)
	tgNotifier := formresponse.NewTelegramNotifier(telegramCfg.BotToken, telegramCfg.ChatID)
	dcNotifier := formresponse.NewDiscordNotifier(discordCfg.WebhookURL)

	// Assemble active notifiers; nil (disabled) adapters are filtered out
	notifiers := formresponse.NewNotifiers(tgNotifier, dcNotifier)

	// Incoming adapter
	formHandler := formresponse.NewFormHandler(notifiers)

	// ── feedback feature ──────────────────────────────────────────────────────
	// Shared infrastructure
	pool, err := db.NewPool(dbCfg)
	if err != nil {
		return nil, err
	}

	// Outgoing adapter (implements the Repository outgoing port)
	repo, err := feedback.NewRepository(pool)
	if err != nil {
		return nil, err
	}

	// Application service (implements the Service incoming port)
	svc := feedback.NewService(repo)

	// Incoming adapter
	fbHandler := feedback.NewHandler(svc)

	// ── Application ──────────────────────────────────────────────────────────
	return app.New(serverCfg, formHandler, fbHandler), nil
}
