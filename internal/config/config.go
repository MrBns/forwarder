package config

import (
	"os"
	"strings"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	Port           string
	AllowedOrigins []string

	// Telegram
	TelegramBotToken string
	TelegramChatID   string

	// Discord
	DiscordWebhookURL string

	// Slack
	SlackWebhookURL string
}

// Load reads configuration from environment variables.
// Missing optional fields simply disable the corresponding notifier.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	rawOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string
	if rawOrigins != "" {
		for _, o := range strings.Split(rawOrigins, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
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
	}
}
