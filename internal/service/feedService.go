package service

import (
	"Forum/internal/domain"
	"Forum/internal/repository"
	"context"
)

type FeedService interface {
	ListHomeFeed(ctx context.Context, userID uint, page, pageSize int) ([]*domain.FeedResponse, int64, error)
}

type feedServiceImpl struct {
	feedRepo repository.FeedRepo
}

func NewFeedService(feedRepo repository.FeedRepo) FeedService {
	return &feedServiceImpl{feedRepo: feedRepo}
}

func (f *feedServiceImpl) ListHomeFeed(ctx context.Context, userID uint, page, pageSize int) ([]*domain.FeedResponse, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	posts, totalCount, err := f.feedRepo.ListHomeFeed(userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	var res []*domain.FeedResponse
	for _, post := range posts {
		res = append(res, &domain.FeedResponse{
			ID:        post.ID,
			UserID:    post.UserID,
			UserName:  post.User.UserName,
			Title:     post.Title,
			Content:   post.Content,
			CreatedAt: post.CreatedAt,
			LikeCount: post.LikeCount,
		})
	}
	return res, totalCount, nil
}
