package repository

import (
	"Forum/internal/domain"

	"gorm.io/gorm"
)

type FeedRepo interface {
	// ListHomeFeed get my feed
	ListHomeFeed(userID uint, limit, offset int) ([]domain.Post, int64, error)
}

type feedRepo struct {
	db *gorm.DB
}

func NewFeedRepo(db *gorm.DB) FeedRepo {
	return &feedRepo{db: db}
}

func (f *feedRepo) ListHomeFeed(userID uint, limit, offset int) ([]domain.Post, int64, error) {
	var posts []domain.Post
	var totalCount int64

	// subquery: my followee (followee_id)
	followeeQry := f.db.Model(&domain.Follow{}).
		Select("followee_id").
		Where("follower_id = ?", userID)

	// main query: posts.user_id IN (followeeQry) OR posts.user_id = userID
	query := f.db.Model(&domain.Post{}).
		Where("status = ?", "published").
		Where(f.db.Where("user_id IN (?)", followeeQry).
			Or("user_id = ?", userID))

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").
		Order("created_at desc, id desc").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, totalCount, nil
}
