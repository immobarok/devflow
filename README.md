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
# Health check
curl http://localhost:8080/api/v1/health

# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "alex@devflow.io", "password": "Password123!", "first_name": "Alex", "last_name": "Rivers"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "alex@devflow.io", "password": "Password123!"}'

# Access protected profile
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

## API Endpoints

| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/health` | Service health status | No |
| `POST` | `/api/v1/auth/register` | Register new user account | No |
| `POST` | `/api/v1/auth/login` | Authenticate user & receive tokens | No |
| `POST` | `/api/v1/auth/refresh` | Rotate and issue fresh access/refresh tokens | No |
| `POST` | `/api/v1/auth/logout` | Revoke refresh token and end session | No |
| `POST` | `/api/v1/auth/forgot-password` | Request 6-digit OTP for password reset | No |
| `POST` | `/api/v1/auth/verify-otp` | Verify OTP and receive temporary reset token | No |
| `POST` | `/api/v1/auth/reset-password` | Reset password using verified reset token | No |
| `GET` | `/api/v1/auth/me` | Fetch authenticated user profile | Yes (Bearer JWT) |
| `POST` | `/api/v1/organizations` | Create workspace (creator becomes owner) | Yes (Bearer JWT) |
| `GET` | `/api/v1/organizations` | List current user's workspaces | Yes (Bearer JWT) |
| `GET` | `/api/v1/organizations/:id` | Get organization details | Yes (Viewer+) |
| `GET` | `/api/v1/organizations/slug/:slug` | Get organization by slug | Yes (Bearer JWT) |
| `PATCH` | `/api/v1/organizations/:id` | Update organization profile | Yes (Admin+) |
| `DELETE` | `/api/v1/organizations/:id` | Delete organization | Yes (Owner) |
| `GET` | `/api/v1/organizations/:id/members` | List members with roles | Yes (Viewer+) |
| `POST` | `/api/v1/organizations/:id/invitations` | Invite teammate with role | Yes (Admin+) |
| `PATCH` | `/api/v1/organizations/:id/members/:member_id` | Update member role | Yes (Admin+) |
| `DELETE` | `/api/v1/organizations/:id/members/:member_id` | Remove member | Yes (Admin+) |
| `POST` | `/api/v1/invitations/accept` | Accept workspace invitation | Yes (Bearer JWT) |
| `POST` | `/api/v1/organizations/:id/projects` | Create project within workspace | Yes (Admin+) |
| `GET` | `/api/v1/organizations/:id/projects` | List projects in workspace | Yes (Viewer+) |
| `GET` | `/api/v1/projects/:id` | Get project details and user's role | Yes (Bearer JWT) |
| `PATCH` | `/api/v1/projects/:id` | Update project name, status, description | Yes (Bearer JWT) |
| `DELETE` | `/api/v1/projects/:id` | Delete project | Yes (Bearer JWT) |
| `GET` | `/api/v1/projects/:id/members` | List members of a project | Yes (Bearer JWT) |
| `POST` | `/api/v1/projects/:id/members` | Add user to project | Yes (Bearer JWT) |
| `PATCH` | `/api/v1/projects/:id/members/:user_id` | Update member project role | Yes (Bearer JWT) |
| `DELETE` | `/api/v1/projects/:id/members/:user_id` | Remove user from project | Yes (Bearer JWT) |
| `POST` | `/api/v1/projects/:id/tasks` | Create task within a project (auto-numbered e.g. DEV-1) | Yes (Project Member) |
| `GET` | `/api/v1/projects/:id/tasks` | List & filter project tasks (status, priority, search, pagination) | Yes (Bearer JWT) |
| `GET` | `/api/v1/tasks/:id` | Get task details with assignee & reporter | Yes (Bearer JWT) |
| `PATCH` | `/api/v1/tasks/:id` | Update task status, priority, type, assignee, etc. | Yes (Bearer JWT) |
| `DELETE` | `/api/v1/tasks/:id` | Delete task | Yes (Bearer JWT) |
| `POST` | `/api/v1/tasks/:id/comments` | Add comment to a task | Yes (Bearer JWT) |
| `GET` | `/api/v1/tasks/:id/comments` | List all comments on a task with author profiles | Yes (Bearer JWT) |
| `PATCH` | `/api/v1/comments/:id` | Update comment content (author only) | Yes (Bearer JWT) |
| `DELETE` | `/api/v1/comments/:id` | Delete comment (author only) | Yes (Bearer JWT) |
| `GET` | `/api/v1/projects/:id/activities` | List activity audit logs for a project | Yes (Bearer JWT) |
| `GET` | `/api/v1/tasks/:id/activities` | List activity audit logs for a specific task | Yes (Bearer JWT) |
| `GET` | `/api/v1/notifications` | List user notifications (paginated) | Yes (Bearer JWT) |
| `GET` | `/api/v1/notifications/unread-count` | Get unread notifications count (partial index) | Yes (Bearer JWT) |
| `PATCH` | `/api/v1/notifications/:id/read` | Mark a notification as read | Yes (Bearer JWT) |
| `POST` | `/api/v1/notifications/read-all` | Mark all user notifications as read | Yes (Bearer JWT) |

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

## License

MIT
# devflow
