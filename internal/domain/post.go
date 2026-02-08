package domain

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Title     string         `gorm:"not null;size:200" json:"title"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Status    string         `gorm:"default:'published';index" json:"status"` // draft/published/archived
	ViewCount int            `gorm:"default:0" json:"view_count"`
	LikeCount int            `gorm:"default:0" json:"like_count"`
	CreatedAt time.Time      `gorm:"not null;index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ----------- DTO -------------

type CreatePostRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required,min=1"`
	Status  string `json:"status" binding:"omitempty,oneof=draft published"`
}

type UpdatePostRequest struct {
	Title   string `json:"title" binding:"omitempty,min=1,max=200"`
	Content string `json:"content" binding:"omitempty,min=1"`
	Status  string `json:"status" binding:"omitempty,oneof=draft published archived"`
}

type PostResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	UserName  string    `json:"user_name"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	ViewCount int       `json:"view_count"`
	LikeCount int       `json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
