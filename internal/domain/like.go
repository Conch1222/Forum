package domain

import (
	"time"

	"gorm.io/gorm"
)

type Like struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"not null;" json:"user_id"`
	TargetID   uint           `gorm:"not null" json:"target_id"`
	TargetType string         `gorm:"not null;size:20" json:"target_type"` // post/comment
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// ----------- DTO -------------

type LikeRequest struct {
	TargetID   uint   `json:"target_id" binding:"required"`
	TargetType string `json:"target_type" binding:"required,oneof=post comment"`
}

type LikeResponse struct {
	IsLiked   bool  `json:"is_liked"`
	LikeCount int64 `json:"like_count"`
}
