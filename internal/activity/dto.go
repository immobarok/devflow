package activity

import (
	"encoding/json"
	"time"
)

type ActorSummary struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type ActivityResponse struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organization_id"`
	ProjectID      string          `json:"project_id"`
	TaskID         *string         `json:"task_id,omitempty"`
	Action         string          `json:"action"`
	Details        json.RawMessage `json:"details" swaggertype:"object"`
	Actor          ActorSummary    `json:"actor"`
	CreatedAt      time.Time       `json:"created_at"`
}

type ActivityListResponse struct {
	Activities []ActivityResponse `json:"activities"`
	Total      int                `json:"total"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
}

func ToActivityResponse(a *ActivityLogWithActor) ActivityResponse {
	return ActivityResponse{
		ID:             a.ID,
		OrganizationID: a.OrganizationID,
		ProjectID:      a.ProjectID,
		TaskID:         a.TaskID,
		Action:         a.Action,
		Details:        a.Details,
		Actor: ActorSummary{
			ID:        a.ActorID,
			Email:     a.ActorEmail,
			FirstName: a.ActorFirst,
			LastName:  a.ActorLast,
			AvatarURL: a.ActorAvatar,
		},
		CreatedAt: a.CreatedAt,
	}
}
