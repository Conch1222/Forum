package repository

import (
	"Forum/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LikeRepo interface {
	Transaction(fn func(tx *gorm.DB) error) error
	WithTx(tx *gorm.DB) LikeRepo

	Create(like *domain.Like) (bool, error)
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

func (l *likeRepo) Transaction(fn func(tx *gorm.DB) error) error {
	return l.db.Transaction(fn)
}

func (l *likeRepo) WithTx(tx *gorm.DB) LikeRepo {
	return &likeRepo{db: tx}
}

// Create if instance exists, do nothing
func (l *likeRepo) Create(like *domain.Like) (bool, error) {
	tx := l.db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(like)

	if tx.Error != nil {
		return false, tx.Error
	}

	return tx.RowsAffected == 1, nil
}

func (l *likeRepo) Delete(userID, targetID uint, targetType string) error {
	return l.db.Where("user_id = ? AND target_id = ? AND target_type = ?",
		userID, targetID, targetType).Delete(&domain.Like{}).Error
}

func (l *likeRepo) IsExist(userID, targetID uint, targetType string) (bool, error) {
	var like domain.Like
	tx := l.db.Model(&domain.Like{}).Where("user_id = ? AND target_id = ? AND target_type = ?",
		userID, targetID, targetType).Limit(1).Find(&like)
	if tx.Error != nil {
		return false, tx.Error
	}

	return tx.RowsAffected > 0, nil
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
