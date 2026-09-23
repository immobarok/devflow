# Go Coding Standards

## General Principles

Write clean, idiomatic Go. Prefer simplicity over cleverness.

Prefer:

```go
func CreateTask(ctx context.Context, task Task) error
```

over unnecessarily complicated abstractions.

---

## Use

- `context.Context` as the first parameter in service and repository functions
- Explicit error handling — never ignore errors without a documented reason
- Meaningful variable names (`taskService` not `ts`)
- Small, focused functions (under ~50 lines)
- Dependency injection via constructor functions
- Interfaces only when they provide real value (e.g., for testing)

## Avoid

- Giant functions or files (split at ~200 lines per file)
- Global mutable state
- Unnecessary interfaces (don't create `TaskServiceInterface` unless you need to mock it)
- Unnecessary design patterns
- Premature optimization
- Unnecessary third-party dependencies

---

## Error Handling

Always handle errors explicitly:

```go
// GOOD
result, err := something()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// BAD — never do this without a documented reason
result, _ := something()
```

### Application Errors

Use custom error types from `pkg/errors/` for business logic errors:

```go
// pkg/errors/errors.go
var (
    ErrNotFound      = errors.New("resource not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrConflict      = errors.New("resource already exists")
    ErrValidation    = errors.New("validation failed")
)
```

### HTTP Status Codes

Map application errors to HTTP status codes in handlers:

```text
400 Bad Request        — malformed request body
401 Unauthorized       — missing or invalid authentication
403 Forbidden          — authenticated but not authorized
404 Not Found          — resource does not exist
409 Conflict           — duplicate resource
422 Unprocessable      — valid syntax but invalid data
429 Too Many Requests  — rate limit exceeded
500 Internal Error     — unexpected server error
```

Do NOT expose internal database errors directly to clients.

---

## Naming Conventions

| Thing | Convention | Example |
|---|---|---|
| Package names | lowercase, single word | `tasks`, `auth`, `middleware` |
| Exported functions | PascalCase | `CreateTask`, `FindByID` |
| Unexported functions | camelCase | `validateInput`, `buildQuery` |
| Constants | PascalCase | `MaxRetries`, `DefaultTimeout` |
| Struct fields | PascalCase | `CreatedAt`, `ProjectID` |
| JSON tags | snake_case | `` `json:"created_at"` `` |
| Database columns | snake_case | `created_at`, `project_id` |
| File names | lowercase with underscores | `task_handler.go` (or `handler.go` if one per package) |

---

## Concurrency

Use Go concurrency intentionally. Not every operation needs a goroutine.

### When to Use

- Background workers (email, notifications)
- Worker pools for job processing
- WebSocket connection management
- Truly asynchronous operations

### Rules

- Always pass `context.Context` for cancellation
- Always handle goroutine cleanup (use `defer`, `sync.WaitGroup`)
- Use channels for communication, mutexes for shared state
- Never launch goroutines that can leak (always have an exit condition)
- Use `sync.WaitGroup` to wait for goroutine completion on shutdown

### When NOT to Use

- Simple CRUD operations
- Sequential request handling
- Operations that are fast enough synchronously
