//go:build wireinject

// This file is the Wire injector. It is excluded from normal compilation by the
// build tag above and is used only by `wire gen` to produce wire_gen.go.
package main

import (
	"github.com/google/wire"

	"github.com/MrBns/form-response/internal/app"
	"github.com/MrBns/form-response/internal/config"
	"github.com/MrBns/form-response/internal/handler"
	"github.com/MrBns/form-response/internal/notifier"
)

// initializeApp is the Wire injector function.
// Wire reads this function's body and generates the wiring code in wire_gen.go.
func initializeApp() (*app.App, error) {
	wire.Build(
		config.ProviderSet,
		notifier.ProviderSet,
		handler.ProviderSet,
		app.ProviderSet,
	)
	return nil, nil
}
