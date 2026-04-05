# FanAPI

A multi-channel LLM & AI generation service aggregation platform. It provides a unified API to proxy multiple third-party AI providers (OpenAI, Claude, etc.) with built-in billing, user, and channel management.

## Features

- **Multi-channel proxy** — Flexible upstream API integration via goja (JavaScript runtime) dynamic scripts for request/response mapping
- **Multi-protocol support** — OpenAI, Claude, and Gemini protocol formats (including SSE streaming)
- **Name-based channel routing** — Set `model` to a channel name; the server resolves the real upstream model automatically. `?channel_id=X` is also supported for backward compatibility
- **LLM chat** — Streaming (SSE) and non-streaming proxy with two-phase billing (pre-deduction + settlement); users can track spend per conversation via `X-Corr-Id` response header
- **Async tasks** — Image, video, and audio generation with async polling, configurable per-request and query timeouts, and automatic refund on failure
- **Billing system** — Multi-dimensional billing models (by token / image / video / audio / custom script), balance management and transaction history
- **Auto refund** — Any task failure (HTTP error, upstream business error, NATS publish failure, timeout) automatically refunds credited balance and writes a refund transaction
- **Card/voucher system** — Admins generate top-up cards; users redeem them for balance
- **User system** — Username + password registration (email optional, for password recovery), JWT login, API Key management
- **Admin panel** — Channel CRUD, key pool management, user recharge, transaction queries, card management, LLM log viewer

## Tech Stack

| Category | Technology |
|----------|-----------|
| Language | Go 1.26 |
| Web Framework | Gin |
| Database | PostgreSQL + xorm |
| Cache | Redis |
| Message Queue | NATS |
| Auth | JWT + API Key |
| Dynamic Scripts | goja (JavaScript) |
| Frontend | Vue 3 + Vite |

## Dependencies

- PostgreSQL (default port 5433)
- Redis (default port 6379)
- NATS (default port 4222)
- SMTP mail service

## Quick Start

### 1. Configuration

```bash
cp config.yaml config.local.yaml
# Edit database, Redis, NATS, SMTP connection settings
```

### 2. Start (development)

```bash
bash scripts/start.sh
```

After startup:
- User portal: `http://localhost:3000`
- Admin panel: `http://localhost:3000/admin`
- API docs: `http://localhost:8080/docs`

### 3. Default accounts

On first startup the database is seeded with:

| Role | Username | Email | Password | Notes |
|------|----------|-------|----------|-------|
| Admin | `admin` | `admin@fanapi.dev` | `Admin@2026!` | Full admin access |
| Test user | `test` | `test@fanapi.dev` | `Test@2026!` | Regular user, for API testing |

> **Change default passwords before deploying to production.**

### 4. Seed data (optional)

```bash
# Pre-built ChatFire channel config
psql -U <user> -d <db> -f scripts/seed_chatfire.sql
```

### 5. Database migration (upgrades only)

New deployments are handled automatically via xorm `Sync2`. For upgrades from older versions:

```bash
psql -U <user> -d <db> -f scripts/migrate_20260405_add_indexes.sql
```

The index migration uses `CONCURRENTLY` and is safe to run on a live database.

## Channel Script System

Each channel can be configured with up to four JavaScript scripts, editable from the admin panel:

| Field | Function | Description |
|-------|----------|-------------|
| `request_script` | `mapRequest(input)` | Transform the platform request into the upstream API format |
| `response_script` | `mapResponse(output)` | Map the upstream sync response to platform format, or extract `upstream_task_id` for async tasks |
| `query_script` | `mapResponse(output)` | Map async poll responses to platform format (`status`: 2=success, 3=failed, other=in progress) |
| `error_script` | `checkError(response)` | Return a non-empty string to trigger a refund and mark the task failed; return `null`/`false` for success |

When `error_script` is not set, the platform uses built-in detection for common error formats (`{"error":{...}}` etc.).

## Billing

1 CNY = 1,000,000 credits

### LLM two-phase billing

| Phase | Timing | Description |
|-------|--------|-------------|
| `hold` (pre-deduct) | Before request | Conservative estimate using max context + max output tokens |
| `settle` (settlement) | After response | Recalculated with actual usage; overpayment refunded or underpayment charged |

Each LLM response includes an `X-Corr-Id` header that maps to the `corr_id` field in billing records.

### Async task billing

| Event | Transaction type | Description |
|-------|-----------------|-------------|
| Task created | `charge` | One-time precise deduction at creation |
| Task failed (any reason) | `refund` | Full automatic refund; `metrics.reason` records the cause |

## API Reference

### Auth (no authentication required)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/send-code` | Send email verification code |
| POST | `/auth/register` | Register a new account |
| POST | `/auth/login` | Login, returns JWT |
| POST | `/auth/forgot-password` | Request password reset code |
| POST | `/auth/reset-password` | Reset password with code |

### User (Bearer JWT or API Key)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/user/profile` | Get user profile |
| GET | `/user/balance` | Get balance |
| GET | `/user/transactions` | Transaction history |
| GET | `/user/channels` | Available channel list (includes `routing_model` field) |
| GET/POST/DELETE | `/user/apikeys` | API Key management |
| PUT | `/user/password` | Change password |
| POST | `/user/bind-email` | Bind email address |
| POST | `/user/cards/redeem` | Redeem a top-up card |
| GET | `/v1/llm-logs` | LLM request logs |

### AI Endpoints (API Key)

Channel routing is done via the `model` field in the request body — set it to the channel **name** (the `routing_model` value from `/user/channels`). The server resolves the real upstream model before forwarding. For backward compatibility, you can also pass `?channel_id=X` as a query parameter.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/chat/completions` | LLM chat — OpenAI-compatible format (SSE supported) |
| POST | `/v1/messages` | LLM chat — Claude native format (SSE supported) |
| POST | `/v1/gemini` | LLM chat — Gemini native format (SSE supported) |
| POST | `/v1/image` | Image generation (async) |
| POST | `/v1/video` | Video generation (async) |
| POST | `/v1/audio` | Audio generation (async) |
| GET | `/v1/tasks` | Task list |
| GET | `/v1/tasks/:id` | Task status query |

### Admin (JWT + admin role)

| Method | Path | Description |
|--------|------|-------------|
| CRUD | `/admin/channels` | Channel management |
| CRUD | `/admin/key-pools` | Key pool management |
| GET/POST/DELETE | `/admin/key-pools/:id/keys` | Pool key management |
| GET | `/admin/users` | User list |
| POST | `/admin/users/:id/recharge` | Recharge user balance |
| PUT | `/admin/users/:id/password` | Reset user password |
| GET | `/admin/transactions` | All transaction records |
| GET | `/admin/tasks` | All task queries |
| GET | `/admin/tasks/:id` | Task detail |
| GET | `/admin/stats` | Platform statistics |
| POST | `/admin/cards/generate` | Bulk generate top-up cards |
| GET | `/admin/cards` | Card list |
| DELETE | `/admin/cards/:id` | Delete card |
| GET | `/admin/llm-logs` | LLM request log viewer |
| GET | `/admin/llm-logs/:id` | LLM request log detail |

## Project Structure

```
fanapi/
├── cmd/
│   ├── server/       # HTTP server entry point
│   └── script/       # Script execution entry point
├── internal/
│   ├── billing/      # Billing engine (extractor, pricer)
│   ├── cache/        # Redis cache
│   ├── config/       # Config loading
│   ├── db/           # Database connection
│   ├── handler/      # HTTP route handlers
│   ├── middleware/   # Auth middleware
│   ├── model/        # Data models
│   ├── mq/           # NATS message queue
│   ├── script/       # Async task workers (NATS only, no DB/Redis)
│   ├── service/      # Business logic layer
│   └── taskresult/   # Result processor, batch writer, async poller
├── pkg/
│   └── mailer/       # Email sending
├── web/
│   └── user/         # Frontend (Vue 3 + Vite — user portal + admin panel)
└── scripts/          # Database init & migration scripts
```

## Contributing

1. Fork this repository
2. Create a `feat/xxx` branch
3. Commit your changes
4. Open a Pull Request
