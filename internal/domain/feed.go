package domain

import "time"

// ----------- DTO -------------

type FeedResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	UserName  string    `json:"user_name"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	LikeCount int       `json:"like_count"`
}
