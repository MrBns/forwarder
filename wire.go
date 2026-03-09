//go:build wireinject

// This file is the Wire injector. It is excluded from normal compilation by the
// build tag above and is used only by `wire gen` to produce wire_gen.go.
package main

import (
	"github.com/google/wire"

	"github.com/MrBns/form-response/features/feedback"
	"github.com/MrBns/form-response/features/formresponse"
	"github.com/MrBns/form-response/internal/app"
	"github.com/MrBns/form-response/internal/config"
	"github.com/MrBns/form-response/internal/db"
)

// initializeApp is the Wire injector function.
// Wire reads this function's body, resolves the dependency graph from the
// provider sets, and generates the wiring code in wire_gen.go.
func initializeApp() (*app.App, error) {
	wire.Build(
		config.ProviderSet,   // Load → *Config → TelegramConfig, DiscordConfig, ServerConfig, DBConfig
		db.NewPool,           // DBConfig → *pgxpool.Pool
		formresponse.ProviderSet, // TelegramConfig, DiscordConfig → Notifiers → *FormHandler
		feedback.ProviderSet, // *pgxpool.Pool → Repository → Service → *Handler
		app.ProviderSet,      // ServerConfig, *FormHandler, *feedback.Handler → *App
	)
	return nil, nil
}
