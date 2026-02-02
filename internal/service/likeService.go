package service

import (
	"Forum/internal/domain"
	"Forum/internal/pkg/cache"
	"Forum/internal/repository"
	"context"
	"time"
)

type LikeService interface {
	Toggle(ctx context.Context, userID, targetID uint, targetType string) (*domain.LikeResponse, error)

	// GetStatus get status of like
	GetStatus(ctx context.Context, userID, targetID uint, targetType string) (*domain.LikeResponse, error)
	BatchLiked(ctx context.Context, userID, targetIDs []uint, targetType string) (map[uint]bool, error)
}

type likeService struct {
	likeRepo repository.LikeRepo
	cache    cache.LikeCache
}

func NewLikeService(likeRepo repository.LikeRepo, cache cache.LikeCache) LikeService {
	return &likeService{likeRepo: likeRepo, cache: cache}
}

func (l *likeService) Toggle(ctx context.Context, userID, targetID uint, targetType string) (*domain.LikeResponse, error) {
	err := l.likeRepo.Create(&domain.Like{
		UserID:     userID,
		TargetID:   targetID,
		TargetType: targetType,
	})

	if err == nil {
		// new like
		l.updateCache(ctx, userID, targetID, targetType, true, 1)
		cnt, _ := l.getCount(ctx, targetID, targetType)
		return &domain.LikeResponse{IsLiked: true, LikeCount: cnt}, nil
	}

	if err.Error() == "already liked" {
		// cancel like
		if derr := l.likeRepo.Delete(userID, targetID, targetType); derr != nil {
			return nil, derr
		}
		l.updateCache(ctx, userID, targetID, targetType, false, -1)
		cnt, _ := l.getCount(ctx, targetID, targetType)
		return &domain.LikeResponse{IsLiked: false, LikeCount: cnt}, nil
	}

	return nil, err

}

func (l *likeService) GetStatus(ctx context.Context, userID, targetID uint, targetType string) (*domain.LikeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *likeService) BatchLiked(ctx context.Context, userID, targetIDs []uint, targetType string) (map[uint]bool, error) {
	//TODO implement me
	panic("implement me")
}

func (l *likeService) updateCache(ctx context.Context, userID, targetID uint, targetType string, isLiked bool, delta int64) {
	if l.cache == nil {
		return
	}

	_ = l.cache.IncrCount(ctx, targetType, targetID, delta)

	if isLiked {
		_ = l.cache.AddUserLike(ctx, userID, targetType, targetID)
	} else {
		_ = l.cache.RemoveUserLike(ctx, userID, targetType, targetID)
	}
}

func (l *likeService) getCount(ctx context.Context, targetID uint, targetType string) (int64, error) {
	// get count from cache, if miss, get from DB and write to cache
	if l.cache != nil {
		if v, err := l.cache.GetCount(ctx, targetType, targetID); err == nil {
			return v, nil
		}
	}

	cnt, err := l.likeRepo.Count(targetID, targetType)
	if err != nil {
		return 0, err
	}

	if l.cache != nil {
		_ = l.cache.SetCount(ctx, targetType, targetID, cnt, time.Hour)
	}
	return cnt, nil
}
