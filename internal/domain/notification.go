package domain

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	UserID uint   `gorm:"not null" json:"user_id"`      // who receive the notification
	Type   string `gorm:"not null;size:30" json:"type"` // "post_liked", "comment_created" ...

	EntityType string `gorm:"size:30;index" json:"entity_type"` // "post"/ "comment"
	EntityID   uint   `gorm:"index" json:"entity_id"`

	Content   string         `gorm:"type:text;not null" json:"content"`
	ReadAt    *time.Time     `gorm:"index" json:"read_at,omitempty"`
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
