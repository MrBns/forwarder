package formresponse

import (
	"log"
)

// NewNotifiers assembles the active Notifiers slice from the optional outgoing
// adapters.  Nil adapters (disabled integrations) are silently omitted.
// It logs a warning when no notifier is active.
func NewNotifiers(tg *TelegramNotifier, dc *DiscordNotifier) Notifiers {
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
