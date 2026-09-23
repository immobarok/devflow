# DevFlow

A production-style project management SaaS built with Go, demonstrating scalable backend engineering practices.

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go, Gin, REST API |
| Database | PostgreSQL (via sqlx) |
| Migrations | golang-migrate |
| Cache | Redis |
| Auth | JWT (access + refresh tokens) |
| Real-time | WebSocket, Redis Pub/Sub |
| Infrastructure | Docker, Docker Compose |
| CI/CD | GitHub Actions |
| Frontend | Next.js, TypeScript, Tailwind CSS |

## Architecture

DevFlow follows a **modular monolith** pattern with the **Handler → Service → Repository** layered architecture per domain module.

```
cmd/server/          → Application entry point
internal/            → Domain modules (auth, users, tasks, etc.)
pkg/                 → Shared utility packages
migrations/          → SQL migration files
```

## Getting Started

### Prerequisites

- Go 1.21+
- Make

### Setup

```bash
# Clone the repository
git clone https://github.com/mobarok/devflow.git
cd devflow

# Copy environment config
cp .env.example .env

# Run the server
make run
```

### Verify

```bash
curl http://localhost:8080/api/v1/health
```

Expected response:

```json
{
  "status": "success",
  "data": {
    "service": "devflow",
    "version": "0.1.0",
    "environment": "development"
  }
}
```

## Available Commands

```bash
make run       # Start the development server
make build     # Build the binary
make test      # Run all tests
make lint      # Run linter
make tidy      # Clean up dependencies
make clean     # Remove build artifacts
make help      # Show all commands
```

## Development Phases

| Phase | Status |
|---|---|
| 1. Project initialization | ✅ Complete |
| 2. Database setup | ⬜ Pending |
| 3. Authentication | ⬜ Pending |
| 4. Users & organizations | ⬜ Pending |
| 5. Projects | ⬜ Pending |
| 6. Tasks | ⬜ Pending |
| 7. Comments & activity logs | ⬜ Pending |
| 8. Redis caching | ⬜ Pending |
| 9. Background workers | ⬜ Pending |
| 10. WebSockets | ⬜ Pending |
| 11. Testing | ⬜ Pending |
| 12. Docker | ⬜ Pending |
| 13. CI/CD | ⬜ Pending |
| 14. Frontend | ⬜ Pending |
| 15. Production deployment | ⬜ Pending |

## License

MIT
# devflow
