# forwarder

A unified, platform-agnostic notification forwarding API built with **Go + Chi**.  
Send one HTTP request — have it delivered to every enabled messaging platform simultaneously.  
Collect and persist user feedback to PostgreSQL when you need a paper trail.

---

## Features

| Feature | Endpoint | Storage |
|---------|----------|---------|
| Forward rich notifications | `POST /api/forward` | Telegram · Discord · Slack |
| Submit a simple form | `POST /api/submit` | Telegram · Discord |
| Collect feedback | `POST /api/feedback` | PostgreSQL |
| Read feedback | `GET /api/feedback` | PostgreSQL |
| Liveness probe | `GET /health` | — |

- **Auto-enable** — a platform activates simply by setting its env vars; no code changes required.
- **Partial-failure resilience** — if one platform fails the others still receive the message.
- **Graceful shutdown** — SIGINT / SIGTERM handled cleanly.
- **Optional database** — the feedback feature is skipped at startup when `DATABASE_URL` is unset.

---

## Project Structure

```
cmd/
  api/
    main.go          entry point — HTTP server with graceful shutdown
    container.go     composition root — wires all features (manual DI)
internal/
  feature/
    forwarder/       POST /api/forward — rich message fan-out
      domain.go        Message · Attachment · Result types; Notifier outgoing port
      handler.go       HTTP incoming adapter
      telegram.go      Telegram outgoing adapter
      discord.go       Discord outgoing adapter
      slack.go         Slack outgoing adapter
    submitter/       POST /api/submit — simple form forwarding
      domain.go        Notifier outgoing port + formatting helpers
      handler.go       HTTP incoming adapter
      telegram.go      Telegram outgoing adapter
      discord.go       Discord outgoing adapter
      notifiers.go     NewNotifiers constructor
    feedback/        POST/GET /api/feedback — Postgres persistence
      domain.go        Feedback model; Repository + Service ports
      service.go       Application service
      repository.go    PostgreSQL outgoing adapter
      handler.go       HTTP incoming adapter
shared/
  config/            Flat env-var config loaded once at startup
  db/                Shared pgxpool.Pool constructor
  server/            Chi router factory + HealthCheck handler
```

Each feature is a self-contained hexagonal slice: the **port** (interface) and its **adapters** (concrete implementations) live in the same feature directory.  
The composition root (`cmd/api/container.go`) is the only place that knows about all features.

---

## Quick Start

```bash
# 1. Clone
git clone https://github.com/MrBns/forwarder
cd forwarder

# 2. Configure
cp .env.example .env
# Edit .env and fill in the credentials for the platforms you want.

# 3. Run
go run ./cmd/api
```

---

## Platform Configuration

| Platform | Env vars required |
|----------|-------------------|
| Discord  | `DISCORD_WEBHOOK_URL` |
| Slack    | `SLACK_WEBHOOK_URL` |
| Telegram | `TELEGRAM_BOT_TOKEN` + `TELEGRAM_CHAT_ID` |
| PostgreSQL | `DATABASE_URL` (feedback feature only) |

Omit a variable entirely to disable that platform — no restarts or code changes needed.

---

## API Reference

### `GET /health`
```json
{"status": "ok"}
```

---

### `POST /api/forward`
Fans a rich message out to all enabled platforms (or a targeted subset).

**Request**
```json
{
  "title":       "Deployment complete",
  "description": "v2.4.1 deployed to production.",
  "note":        "Rollback instructions in the runbook.",
  "footer":      "Triggered by CI pipeline",
  "fields":      { "env": "production", "commit": "a1b2c3d" },
  "attachments": [{ "name": "Release notes", "url": "https://example.com/releases/v2.4.1", "type": "link" }],
  "platforms":   ["discord", "slack"]
}
```
Omit `platforms` to send to **all** enabled platforms.

**Response `200 OK`**
```json
{
  "results": [
    { "platform": "discord", "success": true },
    { "platform": "slack",   "success": false, "error": "slack: unexpected status 503" }
  ]
}
```
**Response `503`** — no platforms enabled or none matched `platforms`.

---

### `POST /api/submit`
Forwards a simple key-value form to Telegram / Discord.

**Request**
```json
{ "fields": { "name": "Alice", "email": "alice@example.com", "message": "Hello!" } }
```

**Response `200 OK`** — `{ "success": true, "message": "form submitted successfully" }`  
**Response `400`** — invalid JSON or empty `fields`.  
**Response `500`** — all notifiers failed.

---

### `POST /api/feedback`
Persists a feedback record to PostgreSQL (requires `DATABASE_URL`).

**Request** — same shape as `/api/submit`

**Response `201 Created`** — the persisted `Feedback` object.

---

### `GET /api/feedback`
Returns paginated feedback records, newest first.

```
GET /api/feedback?limit=20&offset=0
```

**Response `200 OK`**
```json
{
  "feedbacks": [
    { "id": "uuid", "fields": { "rating": "5" }, "origin": "https://your-site.com", "created_at": "2026-03-10T01:00:00Z" }
  ]
}
```

---

## Development

```bash
# Run tests
go test ./...

# Build binary
go build -o forwarder ./cmd/api

# Run with live reload (requires air)
air -c .air.toml
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `ALLOWED_ORIGINS` | `*` | Comma-separated CORS origins |
| `DISCORD_WEBHOOK_URL` | — | Discord incoming webhook |
| `SLACK_WEBHOOK_URL` | — | Slack incoming webhook |
| `TELEGRAM_BOT_TOKEN` | — | Telegram Bot API token |
| `TELEGRAM_CHAT_ID` | — | Telegram target chat/channel ID |
| `DATABASE_URL` | — | PostgreSQL DSN (disables feedback feature when absent) |
