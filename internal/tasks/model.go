package tasks

import "time"

type Task struct {
	ID             string     `db:"id" json:"id"`
	ProjectID      string     `db:"project_id" json:"project_id"`
	OrganizationID string     `db:"organization_id" json:"organization_id"`
	TaskNumber     int        `db:"task_number" json:"task_number"`
	Title          string     `db:"title" json:"title"`
	Description    *string    `db:"description" json:"description,omitempty"`
	Status         string     `db:"status" json:"status"`
	Priority       string     `db:"priority" json:"priority"`
	Type           string     `db:"type" json:"type"`
	AssigneeID     *string    `db:"assignee_id" json:"assignee_id,omitempty"`
	ReporterID     *string    `db:"reporter_id" json:"reporter_id,omitempty"`
	DueDate        *time.Time `db:"due_date" json:"due_date,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

type TaskWithDetails struct {
	Task
	ProjectKey     string  `db:"project_key" json:"project_key"`
	ProjectName    string  `db:"project_name" json:"project_name"`
	AssigneeEmail  *string `db:"assignee_email" json:"assignee_email,omitempty"`
	AssigneeFirst  *string `db:"assignee_first_name" json:"assignee_first_name,omitempty"`
	AssigneeLast   *string `db:"assignee_last_name" json:"assignee_last_name,omitempty"`
	AssigneeAvatar *string `db:"assignee_avatar_url" json:"assignee_avatar_url,omitempty"`
	ReporterEmail  *string `db:"reporter_email" json:"reporter_email,omitempty"`
	ReporterFirst  *string `db:"reporter_first_name" json:"reporter_first_name,omitempty"`
	ReporterLast   *string `db:"reporter_last_name" json:"reporter_last_name,omitempty"`
}
