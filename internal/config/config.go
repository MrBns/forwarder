package config

import (
	"os"
	"strings"

	"github.com/google/wire"
)

// ProviderSet is the Wire provider set for the config package.
var ProviderSet = wire.NewSet(
	Load,
	ProvideTelegramConfig,
	ProvideDiscordConfig,
	ProvideServerConfig,
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Port is the HTTP server listen port (default: 8080).
	Port string

	// AllowedOrigins is a comma-separated list of origins permitted to call the API.
	// Defaults to "*" when not set.
	AllowedOrigins []string

	// TelegramBotToken is the Telegram Bot API token (e.g. "123456:ABC-DEF...").
	// Leave empty to disable Telegram notifications.
	TelegramBotToken string

	// TelegramChatID is the target Telegram chat / channel ID.
	TelegramChatID string

	// DiscordWebhookURL is the full Discord webhook URL.
	// Leave empty to disable Discord notifications.
	DiscordWebhookURL string
}

// TelegramConfig holds the Telegram-specific configuration slice.
type TelegramConfig struct {
	BotToken string
	ChatID   string
}

// DiscordConfig holds the Discord-specific configuration slice.
type DiscordConfig struct {
	WebhookURL string
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port           string
	AllowedOrigins []string
}

// ProvideTelegramConfig extracts TelegramConfig from Config.
func ProvideTelegramConfig(cfg *Config) TelegramConfig {
	return TelegramConfig{BotToken: cfg.TelegramBotToken, ChatID: cfg.TelegramChatID}
}

// ProvideDiscordConfig extracts DiscordConfig from Config.
func ProvideDiscordConfig(cfg *Config) DiscordConfig {
	return DiscordConfig{WebhookURL: cfg.DiscordWebhookURL}
}

// ProvideServerConfig extracts ServerConfig from Config.
func ProvideServerConfig(cfg *Config) ServerConfig {
	return ServerConfig{Port: cfg.Port, AllowedOrigins: cfg.AllowedOrigins}
}

// Load reads configuration from environment variables.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	allowedOrigins := []string{"*"}
	if raw := os.Getenv("ALLOWED_ORIGINS"); raw != "" {
		parts := strings.Split(raw, ",")
		allowedOrigins = make([]string, 0, len(parts))
		for _, o := range parts {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	return &Config{
		Port:              port,
		AllowedOrigins:    allowedOrigins,
		TelegramBotToken:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:    os.Getenv("TELEGRAM_CHAT_ID"),
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
	}
}
