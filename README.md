# form-response

A lightweight Go + Chi web API that accepts HTML form submissions and forwards them to **Telegram** and/or **Discord** — no database required.

---

## Features

- **CORS-restricted** — only origins you trust can call the API.
- **Telegram** forwarding via Bot API (optional).
- **Discord** forwarding via incoming webhook (optional).
- Both, one, or neither notifier can be enabled at runtime via environment variables.
- No database; the server is purely a forwarding layer.

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

| Variable              | Required | Description                                                   |
| --------------------- | -------- | ------------------------------------------------------------- |
| `PORT`                | No       | HTTP listen port (default `8080`)                             |
| `ALLOWED_ORIGINS`     | No       | Comma-separated allowed CORS origins (default `*`)            |
| `TELEGRAM_BOT_TOKEN`  | No       | Telegram Bot API token — omit to disable Telegram             |
| `TELEGRAM_CHAT_ID`    | No       | Telegram target chat / channel ID                             |
| `DISCORD_WEBHOOK_URL` | No       | Discord incoming webhook URL — omit to disable Discord        |

### 3. Run

```bash
# Export variables (or use a tool like direnv / Docker --env-file)
export TELEGRAM_BOT_TOKEN=123456:ABC...
export TELEGRAM_CHAT_ID=-1001234567890
export ALLOWED_ORIGINS=https://your-site.com

./form-response
```

---

## API

### `POST /api/submit`

Submit a form response. The server forwards the payload to all configured notifiers.

**Request body (JSON):**

```json
{
  "fields": {
    "name": "Alice",
    "email": "alice@example.com",
    "message": "Hello!"
  }
}
```

**Success response (`200 OK`):**

```json
{ "success": true, "message": "form submitted successfully" }
```

**Error responses:**

| Status | Reason                              |
| ------ | ----------------------------------- |
| `400`  | Invalid JSON body or empty `fields` |
| `500`  | All configured notifiers failed     |

### `GET /health`

Returns `{ "status": "ok" }` — useful for load-balancer / uptime checks.

---

## Calling from a website

```js
fetch("https://your-api-host/api/submit", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    fields: {
      name:    document.getElementById("name").value,
      email:   document.getElementById("email").value,
      message: document.getElementById("message").value,
    },
  }),
});
```

---

## Development

```bash
# Run tests
go test ./...

# Run locally with a .env file
source .env && go run .
```

---

## Stack

- [Go](https://go.dev/) 1.21+
- [Chi](https://github.com/go-chi/chi) router
- [go-chi/cors](https://github.com/go-chi/cors) middleware
