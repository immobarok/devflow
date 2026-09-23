# Security Rules

## Authentication

### Requirements

- Hash passwords with `bcrypt` (cost factor 12+)
- JWT access tokens (short-lived: 15 minutes)
- Refresh tokens (long-lived: 7 days, stored in database, rotated on use)
- Auth middleware on all protected routes
- Proper logout: revoke refresh token in database

### Never

- Store plaintext passwords
- Hard-code JWT secrets (use environment variables)
- Put sensitive data in JWT payload (only user ID and expiry)
- Trust client-side authorization alone

---

## Authorization — RBAC

Implement organization-level role-based access control.

### Roles

| Role | Permissions |
|---|---|
| `owner` | Full access: manage org, members, projects, tasks, billing |
| `admin` | Manage members (except owner), projects, tasks |
| `developer` | Create/edit tasks and comments, view projects |
| `viewer` | Read-only access to projects and tasks |

### Enforcement Rules

- Every protected endpoint must check the user's role in the relevant organization.
- Authorization is enforced in **middleware** or **service layer**, never only in the handler.
- Resources are always scoped: a user can only access resources within organizations they belong to.
- The RBAC middleware receives the required role and checks `organization_members`.

### Example Middleware

```go
// Require at least "developer" role in the organization
router.PATCH("/tasks/:id", middleware.Auth(), middleware.RequireRole("developer"), taskHandler.Update)
```

---

## General Security Practices

### Always Consider

- **SQL injection**: Use parameterized queries (sqlx handles this). Never concatenate user input into SQL.
- **XSS**: Sanitize output. The API returns JSON so this mainly applies to the frontend.
- **CORS**: Configure allowed origins explicitly. Do not use `*` in production.
- **Rate limiting**: Apply rate limits on auth endpoints (login, register) to prevent brute force.
- **Input validation**: Validate all request bodies, query parameters, and path parameters.
- **Request size limits**: Set max request body size (e.g., 1MB default).

### Never Commit

```text
.env
passwords
API keys
JWT secrets
database credentials
private keys
```

Always provide `.env.example` with placeholder values instead.

### Secret Management

All secrets come from environment variables:

```text
DATABASE_URL=postgres://user:pass@localhost:5432/devflow?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-here
JWT_REFRESH_SECRET=your-refresh-secret-here
```
