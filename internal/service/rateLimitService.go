package service

import (
	"Forum/internal/domain"
	"Forum/internal/repository"
	"context"
	"fmt"
)

type RateLimitService interface {
	Check(ctx context.Context, userID uint, rule domain.RateLimitRule) (*domain.RateLimitResult, error)
}

type rateLimitServiceImpl struct {
	rateLimiterRepo repository.RateLimiterRepo
}

func NewRateLimitService(rateLimiterRepo repository.RateLimiterRepo) RateLimitService {
	return &rateLimitServiceImpl{rateLimiterRepo: rateLimiterRepo}
}

func (r *rateLimitServiceImpl) Check(ctx context.Context, userID uint, rule domain.RateLimitRule) (*domain.RateLimitResult, error) {
	// redis key: rl:{action}:{userID}, ex: rl:create_post:123
	key := fmt.Sprintf("rl:%s:%d", rule.Action, userID)
	return r.rateLimiterRepo.Allow(ctx, key, rule.Limit, rule.Window)
}
