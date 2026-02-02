package cache

import "fmt"

const (
	// KeyLikeCount read like count (string), ex: like:count:post:123 -> 42
	KeyLikeCount = "like:count:%s:%d"

	// KeyUserLikes user already like (set), ex: user:likes:123:post -> Set{5, 8, 12}
	KeyUserLikes = "user:likes:%d:%s"
)

func LikeCountKey(targetType string, targetID uint) string {
	return fmt.Sprintf(KeyLikeCount, targetType, targetID)
}

func UserLikesKey(userID uint, targetType string) string {
	return fmt.Sprintf(KeyUserLikes, userID, targetType)
}
