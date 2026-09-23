# API Design Rules

## Base Path

All API endpoints use the versioned prefix:

```text
/api/v1/...
```

---

## REST Conventions

Use correct HTTP methods:

| Method | Purpose | Example |
|---|---|---|
| GET | Read resource(s) | `GET /api/v1/tasks` |
| POST | Create resource | `POST /api/v1/tasks` |
| PATCH | Partial update | `PATCH /api/v1/tasks/:id` |
| DELETE | Delete resource | `DELETE /api/v1/tasks/:id` |

Do NOT use PUT unless full resource replacement is intended.

---

## Standard Response Format

All API responses MUST use this consistent JSON structure:

### Success Response

```json
{
  "status": "success",
  "data": {
    "id": "uuid-here",
    "title": "Task title"
  }
}
```

### Success Response with Pagination

```json
{
  "status": "success",
  "data": [ ... ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

### Error Response

```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Task title is required",
    "details": [
      { "field": "title", "message": "must not be empty" }
    ]
  }
}
```

### Response Helpers

Implement response helpers in `pkg/response/response.go`:

```go
func Success(c *gin.Context, status int, data interface{})
func SuccessWithMeta(c *gin.Context, status int, data interface{}, meta Meta)
func Error(c *gin.Context, status int, code string, message string)
func ValidationError(c *gin.Context, details []FieldError)
```

---

## Endpoint Map

### Auth

```text
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
```

### Users

```text
GET    /api/v1/users/me
PATCH  /api/v1/users/me
```

### Organizations

```text
POST   /api/v1/organizations
GET    /api/v1/organizations
GET    /api/v1/organizations/:id
PATCH  /api/v1/organizations/:id
DELETE /api/v1/organizations/:id
```

### Organization Members

```text
POST   /api/v1/organizations/:id/members
GET    /api/v1/organizations/:id/members
PATCH  /api/v1/organizations/:id/members/:userId
DELETE /api/v1/organizations/:id/members/:userId
```

### Projects (scoped to organization)

```text
POST   /api/v1/organizations/:orgId/projects
GET    /api/v1/organizations/:orgId/projects
GET    /api/v1/projects/:id
PATCH  /api/v1/projects/:id
DELETE /api/v1/projects/:id
```

### Tasks (scoped to project)

```text
POST   /api/v1/projects/:projectId/tasks
GET    /api/v1/projects/:projectId/tasks
GET    /api/v1/tasks/:id
PATCH  /api/v1/tasks/:id
DELETE /api/v1/tasks/:id
```

### Comments (scoped to task)

```text
POST   /api/v1/tasks/:taskId/comments
GET    /api/v1/tasks/:taskId/comments
PATCH  /api/v1/comments/:id
DELETE /api/v1/comments/:id
```

### Activity Logs

```text
GET    /api/v1/organizations/:orgId/activity
GET    /api/v1/projects/:projectId/activity
```

### Notifications

```text
GET    /api/v1/notifications
PATCH  /api/v1/notifications/:id/read
POST   /api/v1/notifications/read-all
```

---

## Query Parameters

Support these on list endpoints:

```text
?page=1&per_page=20        — Pagination
?sort=created_at&order=desc — Sorting
?status=in_progress         — Filtering
?search=keyword             — Search
```

### Defaults

- `page`: 1
- `per_page`: 20 (max: 100)
- `order`: desc
- `sort`: created_at

---

## Request Validation

Validate all incoming requests using struct tags:

```go
type CreateTaskRequest struct {
    Title       string `json:"title" binding:"required,min=1,max=255"`
    Description string `json:"description" binding:"max=5000"`
    Priority    string `json:"priority" binding:"required,oneof=low medium high critical"`
    AssigneeID  string `json:"assignee_id" binding:"omitempty,uuid"`
}
```

Return `422 Unprocessable Entity` with field-level validation errors.
