package config_test

import (
	"os"
	"testing"

	"github.com/MrBns/form-response/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("ALLOWED_ORIGINS")
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("TELEGRAM_CHAT_ID")
	os.Unsetenv("DISCORD_WEBHOOK_URL")

	cfg := config.Load()

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "*" {
		t.Errorf("AllowedOrigins = %v, want [*]", cfg.AllowedOrigins)
	}
	if cfg.TelegramBotToken != "" {
		t.Errorf("TelegramBotToken = %q, want empty", cfg.TelegramBotToken)
	}
	if cfg.TelegramChatID != "" {
		t.Errorf("TelegramChatID = %q, want empty", cfg.TelegramChatID)
	}
	if cfg.DiscordWebhookURL != "" {
		t.Errorf("DiscordWebhookURL = %q, want empty", cfg.DiscordWebhookURL)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("ALLOWED_ORIGINS", "https://example.com, https://other.com")
	t.Setenv("TELEGRAM_BOT_TOKEN", "mytoken")
	t.Setenv("TELEGRAM_CHAT_ID", "12345")
	t.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/api/webhooks/test")

	cfg := config.Load()

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9090")
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("AllowedOrigins len = %d, want 2", len(cfg.AllowedOrigins))
	} else {
		if cfg.AllowedOrigins[0] != "https://example.com" {
			t.Errorf("AllowedOrigins[0] = %q", cfg.AllowedOrigins[0])
		}
		if cfg.AllowedOrigins[1] != "https://other.com" {
			t.Errorf("AllowedOrigins[1] = %q", cfg.AllowedOrigins[1])
		}
	}
	if cfg.TelegramBotToken != "mytoken" {
		t.Errorf("TelegramBotToken = %q", cfg.TelegramBotToken)
	}
	if cfg.TelegramChatID != "12345" {
		t.Errorf("TelegramChatID = %q", cfg.TelegramChatID)
	}
	if cfg.DiscordWebhookURL != "https://discord.com/api/webhooks/test" {
		t.Errorf("DiscordWebhookURL = %q", cfg.DiscordWebhookURL)
	}
}
