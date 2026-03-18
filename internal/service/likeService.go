package service

import (
	"Forum/internal/domain"
	"Forum/internal/metrics"
	"Forum/internal/pkg/cache"
	"Forum/internal/repository"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
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
	var resp *domain.LikeResponse
	var action string

	// start Transaction
	err := l.likeRepo.Transaction(func(tx *gorm.DB) error {
		likeRepo := l.likeRepo.WithTx(tx)
		postRepo := l.postRepo.WithTx(tx)
		commentRepo := l.commentRepo.WithTx(tx)

		isCreated, createErr := likeRepo.Create(&domain.Like{
			UserID:     userID,
			TargetID:   targetID,
			TargetType: targetType,
		})

		if createErr != nil {
			return createErr
		}

		// new like
		if isCreated {
			switch targetType {
			// add like count
			case "post":
				if err := postRepo.IncrementLikeCount(targetID, 1); err != nil {
					return err
				}
			case "comment":
				if err := commentRepo.IncrementLikeCount(targetID, 1); err != nil {
					return err
				}
			default:
				return fmt.Errorf("invalid target type")
			}

			cnt, err := likeRepo.Count(targetID, targetType)
			if err != nil {
				return err
			}

			resp = &domain.LikeResponse{
				IsLiked:   true,
				LikeCount: cnt,
			}

			action = "liked" // for prometheus
			return nil
		}

		// cancel like
		if !isCreated {
			if dErr := likeRepo.Delete(userID, targetID, targetType); dErr != nil {
				return dErr
			}

			switch targetType {
			// minus like count
			case "post":
				if err := postRepo.IncrementLikeCount(targetID, -1); err != nil {
					return err
				}
			case "comment":
				if err := commentRepo.IncrementLikeCount(targetID, -1); err != nil {
					return err
				}
			default:
				return fmt.Errorf("invalid target type")
			}

			cnt, err := likeRepo.Count(targetID, targetType)
			if err != nil {
				return err
			}

			resp = &domain.LikeResponse{
				IsLiked:   false,
				LikeCount: cnt,
			}

			action = "unliked" // for prometheus
			return nil
		}

		return createErr
	})

	if err != nil {
		return nil, err
	}

	// transaction committed, invalidate cache
	if l.cache != nil {
		_ = l.cache.DelCount(ctx, targetType, targetID)
		_ = l.cache.DeleteUserLikes(ctx, userID, targetType)
	}

	if action != "" {
		metrics.LikeToggleTotal.WithLabelValues(targetType, action).Inc() // for prometheus
	}

	return resp, err
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
