package formresponse

import (
	"log"

	"github.com/google/wire"

	"github.com/MrBns/form-response/internal/config"
)

// ProviderSet is the Wire provider set for the formresponse feature.
// It wires config → outgoing adapters → Notifiers → incoming HTTP adapter.
var ProviderSet = wire.NewSet(
	ProvideTelegramNotifier,
	ProvideDiscordNotifier,
	ProvideNotifiers,
	NewFormHandler,
)

// ProvideTelegramNotifier builds the Telegram outgoing adapter from config.
// Returns nil when credentials are absent, disabling the adapter.
func ProvideTelegramNotifier(cfg config.TelegramConfig) *TelegramNotifier {
	return NewTelegramNotifier(cfg.BotToken, cfg.ChatID)
}

// ProvideDiscordNotifier builds the Discord outgoing adapter from config.
// Returns nil when the webhook URL is absent, disabling the adapter.
func ProvideDiscordNotifier(cfg config.DiscordConfig) *DiscordNotifier {
	return NewDiscordNotifier(cfg.WebhookURL)
}

// ProvideNotifiers assembles the active Notifiers slice.
// Nil adapters (disabled services) are omitted.
func ProvideNotifiers(tg *TelegramNotifier, dc *DiscordNotifier) Notifiers {
	var ns Notifiers
	if tg != nil {
		ns = append(ns, tg)
	}
	if dc != nil {
		ns = append(ns, dc)
	}
	if len(ns) == 0 {
		log.Println("WARNING: no notifier configured – set TELEGRAM_BOT_TOKEN/TELEGRAM_CHAT_ID or DISCORD_WEBHOOK_URL")
	}
	return ns
}
