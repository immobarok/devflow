package tasks

import (
	"fmt"
	"time"
)

type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required,min=2,max=255"`
	Description *string    `json:"description"`
	Status      *string    `json:"status" binding:"omitempty,oneof=backlog todo in_progress in_review done cancelled"`
	Priority    *string    `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
	Type        *string    `json:"type" binding:"omitempty,oneof=task bug feature improvement"`
	AssigneeID  *string    `json:"assignee_id" binding:"omitempty,uuid"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title" binding:"omitempty,min=2,max=255"`
	Description *string    `json:"description"`
	Status      *string    `json:"status" binding:"omitempty,oneof=backlog todo in_progress in_review done cancelled"`
	Priority    *string    `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
	Type        *string    `json:"type" binding:"omitempty,oneof=task bug feature improvement"`
	AssigneeID  *string    `json:"assignee_id" binding:"omitempty,uuid"`
	DueDate     *time.Time `json:"due_date"`
}

type TaskFilterQuery struct {
	Status     string `form:"status"`
	Priority   string `form:"priority"`
	Type       string `form:"type"`
	AssigneeID string `form:"assignee_id"`
	Search     string `form:"search"`
	Limit      int    `form:"limit"`
	Offset     int    `form:"offset"`
}

type UserSummary struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type TaskResponse struct {
	ID             string       `json:"id"`
	ProjectID      string       `json:"project_id"`
	OrganizationID string       `json:"organization_id"`
	Key            string       `json:"key"`
	TaskNumber     int          `json:"task_number"`
	Title          string       `json:"title"`
	Description    *string      `json:"description,omitempty"`
	Status         string       `json:"status"`
	Priority       string       `json:"priority"`
	Type           string       `json:"type"`
	Assignee       *UserSummary `json:"assignee,omitempty"`
	Reporter       *UserSummary `json:"reporter,omitempty"`
	DueDate        *time.Time   `json:"due_date,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type TaskListResponse struct {
	Tasks  []TaskResponse `json:"tasks"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func ToTaskResponse(t *TaskWithDetails) *TaskResponse {
	resp := &TaskResponse{
		ID:             t.ID,
		ProjectID:      t.ProjectID,
		OrganizationID: t.OrganizationID,
		Key:            fmt.Sprintf("%s-%d", t.ProjectKey, t.TaskNumber),
		TaskNumber:     t.TaskNumber,
		Title:          t.Title,
		Description:    t.Description,
		Status:         t.Status,
		Priority:       t.Priority,
		Type:           t.Type,
		DueDate:        t.DueDate,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}

	if t.AssigneeID != nil && t.AssigneeEmail != nil {
		resp.Assignee = &UserSummary{
			ID:        *t.AssigneeID,
			Email:     *t.AssigneeEmail,
			FirstName: *t.AssigneeFirst,
			LastName:  *t.AssigneeLast,
			AvatarURL: t.AssigneeAvatar,
		}
	}

	if t.ReporterID != nil && t.ReporterEmail != nil {
		resp.Reporter = &UserSummary{
			ID:        *t.ReporterID,
			Email:     *t.ReporterEmail,
			FirstName: *t.ReporterFirst,
			LastName:  *t.ReporterLast,
		}
	}

	return resp
}
