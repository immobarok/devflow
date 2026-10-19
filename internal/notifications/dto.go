package notifications

import "time"

type ActorSummary struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type NotificationResponse struct {
	ID         string        `json:"id"`
	Title      string        `json:"title"`
	Message    string        `json:"message"`
	Type       string        `json:"type"`
	EntityType string        `json:"entity_type"`
	EntityID   *string       `json:"entity_id,omitempty"`
	IsRead     bool          `json:"is_read"`
	ReadAt     *time.Time    `json:"read_at,omitempty"`
	Actor      *ActorSummary `json:"actor,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}

type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int                    `json:"total"`
	Unread        int                    `json:"unread"`
	Limit         int                    `json:"limit"`
	Offset        int                    `json:"offset"`
}

type UnreadCountResponse struct {
	UnreadCount int `json:"unread_count"`
}

func ToNotificationResponse(n *NotificationWithActor) NotificationResponse {
	resp := NotificationResponse{
		ID:         n.ID,
		Title:      n.Title,
		Message:    n.Message,
		Type:       n.Type,
		EntityType: n.EntityType,
		EntityID:   n.EntityID,
		IsRead:     n.ReadAt != nil,
		ReadAt:     n.ReadAt,
		CreatedAt:  n.CreatedAt,
	}

	if n.ActorID != nil && n.ActorEmail != nil {
		resp.Actor = &ActorSummary{
			ID:        *n.ActorID,
			Email:     *n.ActorEmail,
			FirstName: *n.ActorFirstName,
			LastName:  *n.ActorLastName,
			AvatarURL: n.ActorAvatar,
		}
	}

	return resp
}
