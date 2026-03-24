package service

import (
	"Forum/internal/domain"
	"Forum/internal/repository"
	"context"
	"strings"
)

type SearchService interface {
	SearchPosts(ctx context.Context, q string, page, pageSize int) ([]domain.PostResponse, int64, error)
}

type searchServiceImpl struct {
	searchRepo repository.SearchRepo
}

func NewSearchService(repo repository.SearchRepo) SearchService {
	return &searchServiceImpl{searchRepo: repo}
}

func (s *searchServiceImpl) SearchPosts(ctx context.Context, q string, page, pageSize int) ([]domain.PostResponse, int64, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []domain.PostResponse{}, 0, nil
	}

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
	return s.searchRepo.SearchPosts(ctx, q, pageSize, offset)
}
