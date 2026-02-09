package service

import (
	"Forum/internal/domain"
	"Forum/internal/repository"
	"context"
	"errors"
)

type FollowService interface {
	Toggle(ctx context.Context, followerID, followeeID uint) (*domain.FollowToggleResponse, error)
	GetStatus(ctx context.Context, followerID, followeeID uint) (bool, error)

	// list: followers / following
	ListFollowing(ctx context.Context, followerID uint, page, pageSize int) ([]domain.FollowListItem, error)
	ListFollowers(ctx context.Context, followeeID uint, page, pageSize int) ([]domain.FollowListItem, error)

	BatchCheckFollowing(ctx context.Context, followerID uint, followeeIDs []uint) (map[uint]bool, error)
}

type followService struct {
	followRepo repository.FollowRepo
}

func NewFollowService(followRepo repository.FollowRepo) FollowService {
	return &followService{followRepo: followRepo}
}

func (f *followService) Toggle(ctx context.Context, followerID, followeeID uint) (*domain.FollowToggleResponse, error) {
	if followerID == followeeID {
		return nil, errors.New("cannot follow yourself")
	}

	err := f.followRepo.Create(&domain.Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
	})

	if err == nil {
		return &domain.FollowToggleResponse{IsFollowing: true}, nil
	}

	if err.Error() == "already followed" {
		if dErr := f.followRepo.Delete(followerID, followeeID); dErr != nil {
			return nil, dErr
		}
		return &domain.FollowToggleResponse{IsFollowing: false}, nil
	}

	return nil, err
}

func (f *followService) GetStatus(ctx context.Context, followerID, followeeID uint) (bool, error) {
	if followerID == followeeID {
		return false, errors.New("cannot follow yourself")
	}

	return f.followRepo.IsExist(followerID, followeeID)
}

func (f *followService) ListFollowing(ctx context.Context, followerID uint, page, pageSize int) ([]domain.FollowListItem, error) {
	limit, offset := normalizePage(page, pageSize)
	follows, err := f.followRepo.ListFollowing(followerID, limit, offset)
	if err != nil {
		return nil, err
	}

	items := make([]domain.FollowListItem, 0, len(follows))
	for _, follow := range follows {
		items = append(items, domain.FollowListItem{
			UserID:     follow.FollowerID,
			UserName:   follow.Followee.UserName,
			FollowedAt: follow.CreatedAt,
		})
	}
	return items, nil
}

func (f *followService) ListFollowers(ctx context.Context, followeeID uint, page, pageSize int) ([]domain.FollowListItem, error) {
	limit, offset := normalizePage(page, pageSize)
	follows, err := f.followRepo.ListFollowers(followeeID, limit, offset)
	if err != nil {
		return nil, err
	}

	items := make([]domain.FollowListItem, 0, len(follows))
	for _, follow := range follows {
		items = append(items, domain.FollowListItem{
			UserID:     follow.FollowerID,
			UserName:   follow.Follower.UserName,
			FollowedAt: follow.CreatedAt,
		})
	}

	return items, nil
}

func (f *followService) BatchCheckFollowing(ctx context.Context, followerID uint, followeeIDs []uint) (map[uint]bool, error) {
	return f.followRepo.BatchCheckFollowing(followerID, followeeIDs)
}

func normalizePage(page, pageSize int) (limit, offset int) {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	limit = pageSize
	offset = (page - 1) * pageSize
	return limit, offset
}
