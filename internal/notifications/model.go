package notifications

import "time"

type Notification struct {
	ID             string     `db:"id" json:"id"`
	UserID         string     `db:"user_id" json:"user_id"`
	ActorID        *string    `db:"actor_id" json:"actor_id,omitempty"`
	OrganizationID *string    `db:"organization_id" json:"organization_id,omitempty"`
	Title          string     `db:"title" json:"title"`
	Message        string     `db:"message" json:"message"`
	Type           string     `db:"type" json:"type"`
	EntityType     string     `db:"entity_type" json:"entity_type"`
	EntityID       *string    `db:"entity_id" json:"entity_id,omitempty"`
	ReadAt         *time.Time `db:"read_at" json:"read_at,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
}

type NotificationWithActor struct {
	Notification
	ActorEmail     *string `db:"actor_email" json:"actor_email,omitempty"`
	ActorFirstName *string `db:"actor_first_name" json:"actor_first_name,omitempty"`
	ActorLastName  *string `db:"actor_last_name" json:"actor_last_name,omitempty"`
	ActorAvatar    *string `db:"actor_avatar_url" json:"actor_avatar_url,omitempty"`
}
