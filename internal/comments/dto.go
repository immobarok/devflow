package comments

import "time"

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

type AuthorSummary struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type CommentResponse struct {
	ID        string        `json:"id"`
	TaskID    string        `json:"task_id"`
	Content   string        `json:"content"`
	Author    AuthorSummary `json:"author"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

func ToCommentResponse(c *CommentWithAuthor) CommentResponse {
	return CommentResponse{
		ID:      c.ID,
		TaskID:  c.TaskID,
		Content: c.Content,
		Author: AuthorSummary{
			ID:        c.UserID,
			Email:     c.AuthorEmail,
			FirstName: c.AuthorFirstName,
			LastName:  c.AuthorLastName,
			AvatarURL: c.AuthorAvatarURL,
		},
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
