# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make run            # go run cmd/api/main.go
make build          # compile to bin/server (CGO_ENABLED=0 GOOS=linux)
make test           # go test ./...
make test-verbose   # go test ./... -v
make test-cover     # generates coverage.html
make swagger        # regenerate docs/ via swag init
make docker-up      # docker compose up --build -d (PostgreSQL + API)
make docker-down    # docker compose down
```

## Architecture

Clean Architecture with manual DI. The dependency chain flows strictly in one direction:

```
main.go → Config → Database → Repository → Service → Handler → Router
```

`cmd/api/main.go` wires everything together manually — no DI framework. Adding a new resource means creating files in `domain`, `repository`, `service`, and `handler`, then wiring in `main.go`.

**Layer responsibilities:**
- `internal/domain/models.go` — all entities, DTOs, enums, and request/response structs in one file
- `internal/repository/` — GORM queries; interface defined alongside implementation
- `internal/service/` — business rules and custom error vars (`ErrScheduleNotFound`, etc.)
- `internal/handler/` — Gin handlers, `router.go`, and integration tests against mocked service
- `pkg/` — stateless utilities: config (godotenv), database (GORM+AutoMigrate), middleware (CORS+logging), response (standard JSON envelope)

**No migration files** — schema is managed entirely via GORM `AutoMigrate` on startup.

## Key Business Rules

- `type=recurring` **requires** `frequency` (daily | weekly | monthly)
- Schedules with `status=sent` cannot be updated or canceled
- Updating contacts does a full replace (delete all, insert new)
- All responses use `{"data": ..., "error": ...}` envelope from `pkg/response`
- Service errors map to HTTP status codes via `handleServiceError()` in the handler

## Testing Pattern

- Unit tests: `internal/service/*_test.go` — mock the repository interface with `testify/mock`
- Integration tests: `internal/handler/*_test.go` — mock the service interface, use `httptest`
- Mocks are manual structs implementing the interface, not generated

## Code Conventions

- Comments in Portuguese, identifiers in English
- Business errors as package-level `var` in `service` package
- Swagger annotations via godoc `@` tags on handlers; regenerate with `make swagger` after changing them
- Timezone: `America/Sao_Paulo` (set in the DB DSN, not in application code)

## Ecosystem Context

This service only **stores** schedules. It does not send messages. The **messaging-officer** service reads these schedules (via **kaizen-secretary** cron worker) and performs the actual WhatsApp delivery.
