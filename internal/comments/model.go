package comments

import "time"

type Comment struct {
	ID        string    `db:"id" json:"id"`
	TaskID    string    `db:"task_id" json:"task_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CommentWithAuthor struct {
	Comment
	AuthorEmail     string  `db:"author_email" json:"author_email"`
	AuthorFirstName string  `db:"author_first_name" json:"author_first_name"`
	AuthorLastName  string  `db:"author_last_name" json:"author_last_name"`
	AuthorAvatarURL *string `db:"author_avatar_url" json:"author_avatar_url,omitempty"`
}
