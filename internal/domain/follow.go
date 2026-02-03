package domain

import (
	"time"

	"gorm.io/gorm"
)

type Follow struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// follower (follows) ->  followee
	FollowerID uint `gorm:"not null" json:"follower_id"`
	FolloweeID uint `gorm:"not null" json:"followee_id"`
	Follower   User `gorm:"foreignKey:FollowerID" json:"follower,omitempty"`
	Followee   User `gorm:"foreignKey:FolloweeID" json:"followee,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
