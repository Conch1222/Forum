package repository

import (
	"Forum/internal/domain"
	"strings"

	"gorm.io/gorm"
)

type SearchRepo interface {
	SearchPosts(q string, limit, offset int) ([]domain.PostResponse, int64, error)
}

type searchRepo struct {
	db *gorm.DB
}

func NewSearchRepo(db *gorm.DB) SearchRepo {
	return &searchRepo{db: db}
}

func (s *searchRepo) SearchPosts(q string, limit, offset int) ([]domain.PostResponse, int64, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []domain.PostResponse{}, 0, nil
	}

	var res []domain.Post
	var count int64

	// plainto_tsquery transform string to tsquery, e.g. 'go redis' -> 'go' & 'redis'
	match := "posts.search_tsv @@ plainto_tsquery('simple', ?)"
	base := s.db.Model(&domain.Post{}).
		Where("status = ?", "published").
		Where(match, q)

	if err := base.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// use rank to calculate and order
	err := base.Select("posts.*, ts_rank(posts.search_tsv, plainto_tsquery('simple', ?)) AS rank", q).
		Preload("User").
		Order("rank desc").
		Order("created_at desc, id desc").
		Limit(limit).
		Offset(offset).
		Find(&res).Error

	var posts []domain.PostResponse
	for _, post := range res {
		posts = append(posts, domain.PostResponse{
			ID:        post.ID,
			UserID:    post.UserID,
			UserName:  post.User.UserName,
			Title:     post.Title,
			Content:   post.Content,
			LikeCount: post.LikeCount,
			CreatedAt: post.CreatedAt,
		})
	}
	return posts, count, err
}
