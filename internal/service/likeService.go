package service

import (
	"Forum/internal/domain"
	"Forum/internal/metrics"
	"Forum/internal/pkg/cache"
	"Forum/internal/repository"
	"context"
	"time"
)

type LikeService interface {
	Toggle(ctx context.Context, userID, targetID uint, targetType string) (*domain.LikeResponse, error)

	// GetStatus get status of like
	GetStatus(ctx context.Context, userID, targetID uint, targetType string) (*domain.LikeResponse, error)
	BatchLiked(ctx context.Context, userID uint, targetIDs []uint, targetType string) (map[uint]bool, error)
}

type likeService struct {
	likeRepo    repository.LikeRepo
	postRepo    repository.PostRepo
	commentRepo repository.CommentRepo
	cache       cache.LikeCache
}

func NewLikeService(likeRepo repository.LikeRepo, postRepo repository.PostRepo, commentRepo repository.CommentRepo, cache cache.LikeCache) LikeService {
	return &likeService{likeRepo: likeRepo, postRepo: postRepo, commentRepo: commentRepo, cache: cache}
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

		// add like count
		if targetType == "post" {
			_ = l.postRepo.IncrementLikeCount(targetID, 1)
		} else if targetType == "comment" {
			_ = l.commentRepo.IncrementLikeCount(targetID, 1)
		}

		metrics.LikeToggleTotal.WithLabelValues(targetType, "liked").Inc() // for prometheus

		cnt, _ := l.getCount(ctx, targetID, targetType)
		return &domain.LikeResponse{IsLiked: true, LikeCount: cnt}, nil
	}

	if err.Error() == "already liked" {
		// cancel like
		if dErr := l.likeRepo.Delete(userID, targetID, targetType); dErr != nil {
			return nil, dErr
		}
		l.updateCache(ctx, userID, targetID, targetType, false, -1)

		// minus like count
		if targetType == "post" {
			_ = l.postRepo.IncrementLikeCount(targetID, -1)
		} else if targetType == "comment" {
			_ = l.commentRepo.IncrementLikeCount(targetID, -1)
		}

		metrics.LikeToggleTotal.WithLabelValues(targetType, "unliked").Inc()

		cnt, _ := l.getCount(ctx, targetID, targetType)
		return &domain.LikeResponse{IsLiked: false, LikeCount: cnt}, nil
	}

	return nil, err

}

func (l *likeService) GetStatus(ctx context.Context, userID, targetID uint, targetType string) (*domain.LikeResponse, error) {
	// get from cache
	isLiked, likedKnown := false, false
	if l.cache != nil {
		if v, err := l.cache.IsLiked(ctx, userID, targetType, targetID); err == nil {
			likedKnown = true
			isLiked = v
		}
	}

	// if not in cache, get from DB
	if !likedKnown {
		v, err := l.likeRepo.IsExist(userID, targetID, targetType)
		if err != nil {
			return nil, err
		}
		isLiked = v
	}

	cnt, err := l.getCount(ctx, targetID, targetType)
	if err != nil {
		return nil, err
	}

	return &domain.LikeResponse{IsLiked: isLiked, LikeCount: cnt}, nil
}

func (l *likeService) BatchLiked(ctx context.Context, userID uint, targetIDs []uint, targetType string) (map[uint]bool, error) {
	if len(targetIDs) == 0 {
		return map[uint]bool{}, nil
	}

	return l.likeRepo.BatchCheckLiked(userID, targetIDs, targetType)
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
