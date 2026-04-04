# FanAPI

A multi-channel LLM & AI generation service aggregation platform. It provides a unified API to proxy multiple third-party AI providers (OpenAI, Claude, etc.) with built-in billing, user, and channel management.

## Features

- **Multi-channel proxy** — Flexible upstream API integration via goja (JavaScript runtime) dynamic scripts for request/response mapping
- **LLM chat** — Streaming (SSE) and non-streaming proxy with two-phase billing (pre-deduction + settlement)
- **Async tasks** — Image, video, and audio generation with async polling for task status
- **Billing system** — Multi-dimensional billing models (by token / image / video / audio / custom script), balance management and transaction history
- **User system** — Email verification code registration, JWT login, API Key management
- **Admin panel** — Channel CRUD, user recharge, transaction queries

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
# Start dependency services via Docker
docker compose -f Dockerfile.dev up -d

# Run the server
go run cmd/server/main.go
```

### 3. Database initialization

```bash
psql -U <user> -d <db> -f scripts/seed_chatfire.sql
```

## API Reference

### Auth (no authentication required)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/send-code` | Send email verification code |
| POST | `/auth/register` | Register a new account |
| POST | `/auth/login` | Login, returns JWT |

### User (Bearer JWT or API Key)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/user/balance` | Get balance |
| GET | `/user/transactions` | Transaction history |
| GET | `/user/channels` | Available channel list |
| GET/POST/DELETE | `/user/apikeys` | API Key management |

### AI Endpoints (API Key)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/llm?channel_id=X` | LLM chat proxy (SSE supported) |
| POST | `/v1/image` | Image generation (async) |
| POST | `/v1/video` | Video generation (async) |
| POST | `/v1/audio` | Audio generation (async) |
| GET | `/v1/tasks` | Task list |
| GET | `/v1/tasks/:id` | Task status query |

### Admin (JWT + admin role)

| Method | Path | Description |
|--------|------|-------------|
| CRUD | `/admin/channels` | Channel management |
| GET | `/admin/users` | User list |
| POST | `/admin/users/:id/recharge` | Recharge user balance |
| GET | `/admin/transactions` | All transaction records |
| GET | `/admin/tasks` | All task queries |

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
│   ├── script/       # Async task workers
│   └── service/      # Business logic layer
├── pkg/
│   └── mailer/       # Email sending
├── web/
│   ├── admin/        # Admin frontend (Vue 3)
│   └── user/         # User frontend (Vue 3)
└── scripts/          # Database init scripts
```

## Contributing

1. Fork this repository
2. Create a `feat/xxx` branch
3. Commit your changes
4. Open a Pull Request
