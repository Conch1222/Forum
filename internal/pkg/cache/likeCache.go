package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type LikeCache interface {
	// Count cache (string/int)
	IncrCount(ctx context.Context, targetType string, targetID uint, delta int64) error
	GetCount(ctx context.Context, targetType string, targetID uint) (int64, error)
	SetCount(ctx context.Context, targetType string, targetID uint, count int64, ttl time.Duration) error
	DelCount(ctx context.Context, targetType string, targetID uint) error

	// User liked cache (set)
	IsLiked(ctx context.Context, userID uint, targetType string, targetID uint) (bool, error)
	AddUserLike(ctx context.Context, userID uint, targetType string, targetID uint) error
	RemoveUserLike(ctx context.Context, userID uint, targetType string, targetID uint) error
	DeleteUserLikes(ctx context.Context, userID uint, targetType string) error
}

type likeCache struct {
	rdb *redis.Client
}

func NewLikeCache(rdb *redis.Client) LikeCache {
	return &likeCache{rdb: rdb}
}

// ----- target like count -----

func (l *likeCache) IncrCount(ctx context.Context, targetType string, targetID uint, delta int64) error {
	key := LikeCountKey(targetType, targetID)
	return l.rdb.IncrBy(ctx, key, delta).Err()
}

func (l *likeCache) GetCount(ctx context.Context, targetType string, targetID uint) (int64, error) {
	key := LikeCountKey(targetType, targetID)
	return l.rdb.Get(ctx, key).Int64()
}

func (l *likeCache) SetCount(ctx context.Context, targetType string, targetID uint, delta int64, ttl time.Duration) error {
	key := LikeCountKey(targetType, targetID)
	return l.rdb.Set(ctx, key, delta, ttl).Err()
}

func (l *likeCache) DelCount(ctx context.Context, targetType string, targetID uint) error {
	key := LikeCountKey(targetType, targetID)
	return l.rdb.Del(ctx, key).Err()
}

// ----- user like -----

func (l *likeCache) IsLiked(ctx context.Context, userID uint, targetType string, targetID uint) (bool, error) {
	key := UserLikesKey(userID, targetType)
	return l.rdb.SIsMember(ctx, key, targetID).Result()
}

func (l *likeCache) AddUserLike(ctx context.Context, userID uint, targetType string, targetID uint) error {
	key := UserLikesKey(userID, targetType)
	return l.rdb.SAdd(ctx, key, targetID).Err()
}

func (l *likeCache) RemoveUserLike(ctx context.Context, userID uint, targetType string, targetID uint) error {
	key := UserLikesKey(userID, targetType)
	return l.rdb.SRem(ctx, key, targetID).Err()
}

func (l *likeCache) DeleteUserLikes(ctx context.Context, userID uint, targetType string) error {
	key := UserLikesKey(userID, targetType)
	return l.rdb.Del(ctx, key).Err()
}
