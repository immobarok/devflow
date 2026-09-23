# Architecture Rules

## Modular Monolith

Start with a modular monolith. Do NOT introduce microservices unless there is a clear technical reason.

Keep modules independent and maintainable. Avoid unnecessary abstraction. Prefer idiomatic Go.

---

## Project Structure

```text
devflow/
├── cmd/
│   └── server/
│       └── main.go                  # Entry point: wire dependencies, start server
│
├── internal/
│   ├── auth/
│   │   ├── handler.go               # Gin HTTP handlers
│   │   ├── service.go               # Business logic
│   │   ├── repository.go            # Database queries
│   │   ├── model.go                 # Domain structs (DB models)
│   │   ├── dto.go                   # Request/Response DTOs
│   │   └── routes.go                # Route registration
│   │
│   ├── users/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── model.go
│   │   ├── dto.go
│   │   └── routes.go
│   │
│   ├── organizations/
│   │   └── ... (same pattern)
│   │
│   ├── projects/
│   │   └── ... (same pattern)
│   │
│   ├── tasks/
│   │   └── ... (same pattern)
│   │
│   ├── comments/
│   │   └── ... (same pattern)
│   │
│   ├── notifications/
│   │   └── ... (same pattern)
│   │
│   ├── activity/
│   │   └── ... (same pattern)
│   │
│   ├── middleware/
│   │   ├── auth.go                  # JWT authentication
│   │   ├── rbac.go                  # Role-based authorization
│   │   ├── ratelimit.go             # Rate limiting
│   │   ├── cors.go                  # CORS configuration
│   │   └── logger.go                # Request logging
│   │
│   ├── workers/
│   │   ├── dispatcher.go            # Worker pool dispatcher
│   │   ├── email.go                 # Email worker
│   │   └── notification.go          # Notification worker
│   │
│   └── websocket/
│       ├── hub.go                   # Connection hub
│       ├── client.go                # Client handler
│       └── handler.go               # HTTP upgrade handler
│
├── pkg/
│   ├── config/
│   │   └── config.go                # App configuration (env loading)
│   ├── database/
│   │   └── postgres.go              # PostgreSQL connection pool
│   ├── cache/
│   │   └── redis.go                 # Redis connection
│   ├── jwt/
│   │   └── jwt.go                   # Token generation/validation
│   ├── logger/
│   │   └── logger.go                # Structured logging
│   ├── response/
│   │   └── response.go              # Standardized API responses
│   ├── validator/
│   │   └── validator.go             # Input validation utilities
│   └── errors/
│       └── errors.go                # Custom application error types
│
├── migrations/
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   └── ...
│
├── docs/
│   └── api.md
│
├── scripts/
│   └── seed.go
│
├── .env.example
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

## Per-Module Pattern: Handler → Service → Repository

Every domain module (`auth/`, `users/`, `tasks/`, etc.) MUST follow this layered pattern:

| File | Responsibility | Depends On |
|---|---|---|
| `handler.go` | Parse HTTP request, call service, return HTTP response | Service |
| `service.go` | Business logic, validation, orchestration | Repository |
| `repository.go` | Raw database queries only, no business logic | Database |
| `model.go` | Domain structs that map to database tables | Nothing |
| `dto.go` | Request/response structs for the API layer | Nothing |
| `routes.go` | Register Gin routes for this module | Handler |

### Example Flow

```text
HTTP Request
  → handler.go (parse request, validate DTO)
    → service.go (apply business rules)
      → repository.go (execute SQL query)
    ← service.go (return result or error)
  ← handler.go (format response, set status code)
HTTP Response
```

### Rules

- Handlers must NOT contain business logic.
- Services must NOT import `gin` or know about HTTP.
- Repositories must NOT contain business logic — only data access.
- Models represent database entities. DTOs represent API contracts. Keep them separate.
- Each module registers its own routes in `routes.go`.

---

## Package Guidelines

### `internal/` — Private application code

- Domain-specific modules that are NOT importable by external projects.
- Each module is self-contained with its own handler, service, repository.

### `pkg/` — Shared utilities

- Reusable packages that could theoretically be imported by other projects.
- Must NOT import from `internal/`.
- Keep these thin: config loading, database connection, JWT helpers, response formatting.

### `cmd/` — Application entry points

- `cmd/server/main.go` wires all dependencies together and starts the HTTP server.
- No business logic in `main.go`. Only dependency injection and server startup.

---

## Dependency Injection

Wire dependencies in `main.go` and pass them down:

```go
// main.go
db := database.Connect(cfg)
taskRepo := tasks.NewRepository(db)
taskService := tasks.NewService(taskRepo)
taskHandler := tasks.NewHandler(taskService)
tasks.RegisterRoutes(router, taskHandler)
```

Do NOT use global variables for dependencies. Do NOT use dependency injection frameworks.
