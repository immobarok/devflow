CREATE TABLE tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    task_number     INTEGER NOT NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    status          VARCHAR(50) NOT NULL DEFAULT 'todo'
                    CHECK (status IN ('backlog', 'todo', 'in_progress', 'in_review', 'done', 'cancelled')),
    priority        VARCHAR(50) NOT NULL DEFAULT 'medium'
                    CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    type            VARCHAR(50) NOT NULL DEFAULT 'task'
                    CHECK (type IN ('task', 'bug', 'feature', 'improvement')),
    assignee_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    reporter_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    due_date        TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (project_id, task_number)
);

CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_org_id ON tasks(organization_id);
CREATE INDEX idx_tasks_project_status ON tasks(project_id, status);
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
