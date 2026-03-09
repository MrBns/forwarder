# form-response

A lightweight **Go + Chi** web API that:

- **Forwards** HTML form submissions to **Telegram** and/or **Discord** (no database).
- **Persists** feedback to **PostgreSQL** and exposes a list endpoint.

Both features share the same codebase, use **Google Wire** for compile-time dependency injection, and follow a **Hexagonal Architecture** with every port adapter living inside its own feature directory.

---

## Project layout

```
features/
  formresponse/          # Form-submission feature (→ Telegram / Discord)
    domain.go            #   Notifier outgoing port + formatting helpers
    telegram.go          #   Telegram outgoing adapter
    discord.go           #   Discord outgoing adapter
    handler.go           #   HTTP incoming adapter  (POST /api/submit)
    providers.go         #   Wire provider set
    *_test.go
  feedback/              # Feedback feature (→ PostgreSQL)
    domain.go            #   Feedback model, Repository & Service ports
    service.go           #   Application service (implements Service)
    repository.go        #   PostgreSQL outgoing adapter (implements Repository)
    handler.go           #   HTTP incoming adapter  (POST/GET /api/feedback)
    providers.go         #   Wire provider set + wire.Bind interface bindings
    *_test.go
internal/
  config/                # Environment-variable configuration + Wire providers
  db/                    # Shared pgxpool.Pool provider
  app/                   # Chi router assembly + App struct
wire.go                  # Wire injector  (build-tagged, not compiled normally)
wire_gen.go              # Wire-generated wiring  (committed, do not edit by hand)
main.go                  # Entry point — calls initializeApp() from wire_gen.go
```

---

## Quick start

### 1. Clone & build

```bash
git clone https://github.com/MrBns/form-response.git
cd form-response
go build -o form-response .
```

### 2. Configure environment

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

| Variable              | Required | Description                                                |
| --------------------- | -------- | ---------------------------------------------------------- |
| `PORT`                | No       | HTTP listen port (default `8080`)                          |
| `ALLOWED_ORIGINS`     | No       | Comma-separated CORS origins (default `*`)                 |
| `TELEGRAM_BOT_TOKEN`  | No       | Telegram Bot API token — omit to disable Telegram          |
| `TELEGRAM_CHAT_ID`    | No       | Target Telegram chat / channel ID                          |
| `DISCORD_WEBHOOK_URL` | No       | Discord incoming webhook URL — omit to disable Discord     |
| `DATABASE_URL`        | **Yes**  | PostgreSQL DSN (`postgres://user:pass@host:5432/db`)       |

### 3. Start Postgres (example with Docker)

```bash
docker run -d --name pg \
  -e POSTGRES_USER=user -e POSTGRES_PASSWORD=pass -e POSTGRES_DB=formresponse \
  -p 5432:5432 postgres:16
```

### 4. Run

```bash
source .env   # or export variables manually
./form-response
```

The `feedbacks` table is created automatically on first startup.

---

## API

### Form submission — `POST /api/submit`

Accepts a form payload and forwards it to all configured notifiers.

```http
POST /api/submit
Content-Type: application/json

{ "fields": { "name": "Alice", "email": "alice@example.com", "message": "Hello!" } }
```

| Status | Meaning                             |
| ------ | ----------------------------------- |
| `200`  | Delivered to at least one notifier  |
| `400`  | Invalid JSON or empty `fields`      |
| `500`  | All notifiers failed                |

---

### Submit feedback — `POST /api/feedback`

Persists a feedback record to PostgreSQL and returns it.

```http
POST /api/feedback
Content-Type: application/json

{ "fields": { "rating": "5", "comment": "Love it!" } }
```

| Status | Meaning               |
| ------ | --------------------- |
| `201`  | Feedback saved        |
| `400`  | Invalid JSON / empty  |
| `500`  | Database error        |

---

### List feedbacks — `GET /api/feedback`

Returns paginated feedbacks, newest first.

```http
GET /api/feedback?limit=20&offset=0
```

Response:

```json
{
  "feedbacks": [
    {
      "id": "uuid",
      "fields": { "rating": "5", "comment": "Love it!" },
      "origin": "https://your-site.com",
      "created_at": "2026-03-09T18:00:00Z"
    }
  ]
}
```

---

### Health check — `GET /health`

Returns `{ "status": "ok" }` — for load-balancer / uptime probes.

---

## Calling from a website

```js
// Form submission (→ Telegram / Discord)
fetch("https://your-api/api/submit", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ fields: { name: "Alice", message: "Hi" } }),
});

// Feedback (→ PostgreSQL)
fetch("https://your-api/api/feedback", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ fields: { rating: "5", comment: "Great!" } }),
});
```

---

## Development

```bash
# Run tests
go test ./...

# Regenerate Wire wiring after changing providers
go run github.com/google/wire/cmd/wire ./...

# Run locally
source .env && go run .
```

---

## Architecture

The project follows **Hexagonal Architecture** (Ports & Adapters):

- **Incoming adapters** (`handler.go` in each feature) translate HTTP requests into calls on the application's incoming port.
- **Outgoing adapters** (`telegram.go`, `discord.go`, `repository.go`) implement the outgoing ports defined in `domain.go`.
- **Google Wire** resolves the full dependency graph at compile time — `wire_gen.go` is generated and committed so the binary can be built without the `wire` tool installed.

---

## Stack

- [Go](https://go.dev/) 1.21+
- [Chi](https://github.com/go-chi/chi) router
- [go-chi/cors](https://github.com/go-chi/cors) middleware
- [Google Wire](https://github.com/google/wire) compile-time DI
- [pgx v5](https://github.com/jackc/pgx) PostgreSQL driver
