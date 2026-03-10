// Package config loads all application configuration from environment variables.
// It is a shared package consumed by the composition root and every feature
// that needs credentials or infrastructure settings.
package config

import (
	"os"
	"strings"
)

// Config holds every configuration value for the application.
// Fields are flat — there are no nested sub-types — so the composition root
// can pick exactly what it needs without intermediate adapters.
type Config struct {
	// Server
	Port           string
	AllowedOrigins []string

	// Telegram — consumed by both the forwarder and submitter features.
	TelegramBotToken string
	TelegramChatID   string

	// Discord — consumed by both the forwarder and submitter features.
	DiscordWebhookURL string

	// Slack — consumed by the forwarder feature.
	SlackWebhookURL string

	// DatabaseURL is the PostgreSQL DSN consumed by the feedback feature.
	// Leave empty to disable the feedback feature at runtime.
	DatabaseURL string
}

// Load reads configuration from environment variables.
// Missing optional fields simply disable the corresponding feature/adapter.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var origins []string
	if raw := os.Getenv("ALLOWED_ORIGINS"); raw != "" {
		for _, o := range strings.Split(raw, ",") {
			if o = strings.TrimSpace(o); o != "" {
				origins = append(origins, o)
			}
		}
	}
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	return &Config{
		Port:              port,
		AllowedOrigins:    origins,
		TelegramBotToken:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:    os.Getenv("TELEGRAM_CHAT_ID"),
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
		SlackWebhookURL:   os.Getenv("SLACK_WEBHOOK_URL"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
	}
}
