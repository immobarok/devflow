# Database Rules

## Technology

- **Database**: PostgreSQL
- **Driver**: `sqlx` (thin wrapper over `database/sql`, close to raw SQL)
- **Migrations**: `golang-migrate` (SQL-based, CLI + library)

### Why sqlx?

- Close to raw SQL — you learn real database skills
- No ORM magic — every query is explicit and debuggable
- Struct scanning — maps rows to Go structs with `db` tags
- Industry standard — most Go teams use either `sqlx` or `database/sql`

### Why golang-migrate?

- Most popular Go migration tool
- SQL-based migrations (you write real SQL, not Go DSL)
- CLI for manual use + library for programmatic use
- Supports up/down migrations for rollback

---

## Schema Design

### Entities

```text
users
organizations
organization_members
projects
project_members
tasks
comments
notifications
activity_logs
refresh_tokens
```

### Conventions

| Convention | Rule |
|---|---|
| Primary keys | UUID (`gen_random_uuid()`) |
| Timestamps | `created_at`, `updated_at` with `DEFAULT NOW()` |
| Soft deletes | `deleted_at TIMESTAMPTZ NULL` where appropriate |
| Foreign keys | Always defined with `ON DELETE` behavior |
| Indexes | On foreign keys, frequently filtered columns, unique constraints |
| Column names | `snake_case` |
| Table names | Plural `snake_case` |
| Enums | Use PostgreSQL `TEXT` with CHECK constraints (not PG enums, they're hard to migrate) |

### Example Migration

```sql
-- 000001_create_users.up.sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      VARCHAR(255) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name  VARCHAR(100) NOT NULL,
    avatar_url TEXT,
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

```sql
-- 000001_create_users.down.sql
DROP TABLE IF EXISTS users;
```

---

## Migration Rules

- Every schema change requires a migration file (up + down).
- Migration files are numbered sequentially: `000001_`, `000002_`, etc.
- Never manually modify a production schema without a migration.
- Down migrations must cleanly reverse the up migration.
- Test both up and down migrations.

---

## Query Rules

- Use parameterized queries. Never concatenate user input into SQL.
- Use `sqlx.Get` for single rows, `sqlx.Select` for multiple rows.
- Use named queries for inserts/updates with many fields.
- Handle `sql.ErrNoRows` explicitly — convert to application-level `ErrNotFound`.

### Example Repository

```go
func (r *TaskRepository) FindByID(ctx context.Context, id uuid.UUID) (*Task, error) {
    var task Task
    err := r.db.GetContext(ctx, &task, "SELECT * FROM tasks WHERE id = $1 AND deleted_at IS NULL", id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, apperrors.ErrNotFound
        }
        return nil, fmt.Errorf("find task by id: %w", err)
    }
    return &task, nil
}
```

---

## Transaction Rules

Use transactions when multiple related operations must succeed or fail together:

```go
func (r *Repository) CreateWithMembers(ctx context.Context, org Organization, ownerID uuid.UUID) error {
    tx, err := r.db.BeginTxx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Insert organization
    _, err = tx.NamedExecContext(ctx, `INSERT INTO organizations (...) VALUES (...)`, org)
    if err != nil {
        return fmt.Errorf("insert organization: %w", err)
    }

    // Insert owner as member
    _, err = tx.ExecContext(ctx, `INSERT INTO organization_members (...) VALUES (...)`, org.ID, ownerID, "owner")
    if err != nil {
        return fmt.Errorf("insert owner member: %w", err)
    }

    return tx.Commit()
}
```

### When to use transactions

- Creating an organization + adding the owner as a member
- Deleting a project + cleaning up related tasks
- Any operation where partial completion would leave bad data

### When NOT to use transactions

- Simple single-row CRUD operations
- Read-only queries
