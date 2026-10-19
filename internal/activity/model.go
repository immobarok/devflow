package activity

import (
	"encoding/json"
	"time"
)

type ActivityLog struct {
	ID             string          `db:"id" json:"id"`
	OrganizationID string          `db:"organization_id" json:"organization_id"`
	ProjectID      string          `db:"project_id" json:"project_id"`
	TaskID         *string         `db:"task_id" json:"task_id,omitempty"`
	ActorID        string          `db:"actor_id" json:"actor_id"`
	Action         string          `db:"action" json:"action"`
	Details        json.RawMessage `db:"details" json:"details"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}

type ActivityLogWithActor struct {
	ActivityLog
	ActorEmail  string  `db:"actor_email" json:"actor_email"`
	ActorFirst  string  `db:"actor_first_name" json:"actor_first_name"`
	ActorLast   string  `db:"actor_last_name" json:"actor_last_name"`
	ActorAvatar *string `db:"actor_avatar_url" json:"actor_avatar_url,omitempty"`
}
