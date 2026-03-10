# forwarder

A unified, platform-agnostic notification forwarding API. Send a single HTTP request and have it delivered to every enabled messaging platform (Discord, Slack, Telegram) simultaneously.

---

## Features

- **Single endpoint** – `POST /api/forward` forwards your message to all enabled platforms in one call.
- **Flexible payload** – supports `title`, `description`, `note`, `footer`, `fields` (key-value map), and `attachments` (files / links).
- **Platform targeting** – optionally choose which platforms receive a specific message via the `platforms` field.
- **Auto-enable** – a platform is enabled simply by setting its credentials in the environment; no code changes needed.
- **Partial-failure resilience** – if one platform fails the others still receive the message; per-platform results are returned.
- **Graceful shutdown** – SIGINT / SIGTERM are handled cleanly.
- **CORS** – configurable allowed origins.

---

## Supported Platforms

| Platform | Required env vars |
|----------|-------------------|
| Discord  | `DISCORD_WEBHOOK_URL` |
| Slack    | `SLACK_WEBHOOK_URL` |
| Telegram | `TELEGRAM_BOT_TOKEN` + `TELEGRAM_CHAT_ID` |

---

## Quick Start

```bash
# 1. Clone
git clone https://github.com/MrBns/forwarder
cd forwarder

# 2. Configure
cp .env.example .env
# Edit .env and fill in the credentials for the platforms you want to enable.

# 3. Run
go run .
# Server starts on :8080
```

---

## API Reference

### `GET /health`

Liveness probe.

**Response `200 OK`**
```json
{"status": "ok"}
```

---

### `POST /api/forward`

Forward a message to all enabled platforms (or a subset).

**Request body** (`application/json`)

| Field | Type | Description |
|-------|------|-------------|
| `title` | `string` | Short heading for the notification |
| `description` | `string` | Main body text |
| `note` | `string` | Optional callout / remark |
| `footer` | `string` | Small metadata shown at the bottom |
| `fields` | `object` | Arbitrary key-value pairs for structured data |
| `attachments` | `array` | Files or links (see below) |
| `platforms` | `array of string` | Target specific platforms (`"discord"`, `"slack"`, `"telegram"`). Omit to send to **all** enabled platforms. |

**Attachment object**

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Display name of the attachment |
| `url` | `string` | URL to the file or resource |
| `type` | `string` | Hint: `"image"`, `"file"`, `"link"`, etc. (optional) |

**Response `200 OK`**
```json
{
  "results": [
    {"platform": "discord",  "success": true},
    {"platform": "slack",    "success": true},
    {"platform": "telegram", "success": false, "error": "telegram: unexpected status 401"}
  ]
}
```

**Response `503 Service Unavailable`** – no platforms are enabled or none matched the requested `platforms` list.

---

## Examples

### Full message to all platforms

```bash
curl -X POST http://localhost:8080/api/forward \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Deployment complete",
    "description": "Version 2.4.1 has been deployed to production.",
    "note": "Rollback instructions available in the runbook.",
    "footer": "Triggered by CI pipeline",
    "fields": {
      "environment": "production",
      "duration":    "3m 12s",
      "commit":      "a1b2c3d"
    },
    "attachments": [
      {"name": "Release notes", "url": "https://example.com/releases/2.4.1", "type": "link"}
    ]
  }'
```

### Target only Discord and Slack

```bash
curl -X POST http://localhost:8080/api/forward \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Alert",
    "description": "High memory usage detected.",
    "platforms": ["discord", "slack"]
  }'
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Port the HTTP server listens on |
| `ALLOWED_ORIGINS` | `*` | Comma-separated CORS allowed origins |
| `DISCORD_WEBHOOK_URL` | — | Discord incoming webhook URL |
| `SLACK_WEBHOOK_URL` | — | Slack incoming webhook URL |
| `TELEGRAM_BOT_TOKEN` | — | Telegram Bot API token |
| `TELEGRAM_CHAT_ID` | — | Telegram target chat / channel ID |

---

## Development

```bash
# Build
go build -o forwarder .

# Run tests
go test ./...

# Run with live reload (requires air)
air
```
