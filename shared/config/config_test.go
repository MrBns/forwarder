package config_test

import (
	"os"
	"testing"

	"github.com/mrbns/forwarder/shared/config"
)

func TestLoad_Defaults(t *testing.T) {
	// Ensure variables are unset.
	for _, k := range []string{"PORT", "ALLOWED_ORIGINS", "TELEGRAM_BOT_TOKEN",
		"TELEGRAM_CHAT_ID", "DISCORD_WEBHOOK_URL", "SLACK_WEBHOOK_URL", "DATABASE_URL"} {
		os.Unsetenv(k)
	}

	cfg := config.Load()

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "*" {
		t.Errorf("AllowedOrigins = %v, want [*]", cfg.AllowedOrigins)
	}
	if cfg.TelegramBotToken != "" || cfg.TelegramChatID != "" {
		t.Error("expected empty Telegram credentials")
	}
	if cfg.DiscordWebhookURL != "" {
		t.Error("expected empty Discord URL")
	}
	if cfg.SlackWebhookURL != "" {
		t.Error("expected empty Slack URL")
	}
	if cfg.DatabaseURL != "" {
		t.Error("expected empty DatabaseURL")
	}
}

func TestLoad_CustomValues(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("ALLOWED_ORIGINS", "https://a.com, https://b.com")
	os.Setenv("TELEGRAM_BOT_TOKEN", "mytoken")
	os.Setenv("TELEGRAM_CHAT_ID", "42")
	os.Setenv("DISCORD_WEBHOOK_URL", "https://discord.example.com/hook")
	os.Setenv("SLACK_WEBHOOK_URL", "https://slack.example.com/hook")
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Cleanup(func() {
		for _, k := range []string{"PORT", "ALLOWED_ORIGINS", "TELEGRAM_BOT_TOKEN",
			"TELEGRAM_CHAT_ID", "DISCORD_WEBHOOK_URL", "SLACK_WEBHOOK_URL", "DATABASE_URL"} {
			os.Unsetenv(k)
		}
	})

	cfg := config.Load()

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090", cfg.Port)
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("AllowedOrigins = %v, want 2 entries", cfg.AllowedOrigins)
	}
	if cfg.TelegramBotToken != "mytoken" {
		t.Errorf("TelegramBotToken = %q, want mytoken", cfg.TelegramBotToken)
	}
	if cfg.SlackWebhookURL != "https://slack.example.com/hook" {
		t.Errorf("SlackWebhookURL = %q", cfg.SlackWebhookURL)
	}
	if cfg.DatabaseURL != "postgres://localhost/test" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
}
