package repository

import (
	"Forum/internal/domain"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type LikeRepo interface {
	Create(like *domain.Like) error
	Delete(userID, targetID uint, targetType string) error
	IsExist(userID, targetID uint, targetType string) (bool, error)
	Count(targetID uint, targetType string) (int64, error) // count total likes
	BatchCheckLiked(userID uint, targetIDs []uint, targetType string) (map[uint]bool, error)
}

type likeRepo struct {
	db *gorm.DB
}

func NewLikeRepo(db *gorm.DB) LikeRepo {
	return &likeRepo{db: db}
}

func (l *likeRepo) Create(like *domain.Like) error {
	err := l.db.Create(like).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "UNIQUE constraint") {
			return errors.New("already liked")
		}
		return err
	}
	return nil
}

func (l *likeRepo) Delete(userID, targetID uint, targetType string) error {
	return l.db.Where("user_id = ? AND target_id = ? AND target_type = ?",
		userID, targetID, targetType).Delete(&domain.Like{}).Error
}

func (l *likeRepo) IsExist(userID, targetID uint, targetType string) (bool, error) {
	var count int64
	err := l.db.Model(&domain.Like{}).Where("user_id = ? AND target_id = ? AND target_type = ?",
		userID, targetID, targetType).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (l *likeRepo) Count(targetID uint, targetType string) (int64, error) {
	var count int64
	err := l.db.Model(&domain.Like{}).Where("target_id = ? AND target_type = ?",
		targetID, targetType).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (l *likeRepo) BatchCheckLiked(userID uint, targetIDs []uint, targetType string) (map[uint]bool, error) {
	var likes []domain.Like
	l.db.Where("user_id = ? AND target_type = ? AND target_id IN ?",
		userID, targetType, targetIDs).Find(&likes)

	res := make(map[uint]bool)
	for _, like := range likes {
		res[like.TargetID] = true
	}
	return res, nil
}
