package repository

import (
	"Forum/internal/domain"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type FollowRepo interface {
	Create(follow *domain.Follow) error
	Delete(followerID, followedID uint) error
	IsExist(followerID, followedID uint) (bool, error)
	ListFollowing(followerID uint, limit, offset int) ([]*domain.Follow, error)     //  list the users I follow
	ListFollowers(followedID uint, limit, offset int) ([]*domain.Follow, error)     // list the followers
	BatchCheckFollowing(followerID uint, followeeIDs []uint) (map[uint]bool, error) // check if I follow multiple users
	CountFollowers(followedID uint) (int64, error)
	CountFollowing(followerID uint) (int64, error)
}

func NewFollowRepo(db *gorm.DB) FollowRepo {
	return &followRepo{db: db}
}

type followRepo struct {
	db *gorm.DB
}

func (f *followRepo) Create(follow *domain.Follow) error {
	err := f.db.Create(follow).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "UNIQUE constraint") {
			return errors.New("already followed")
		}
		return err
	}
	return nil
}

func (f *followRepo) Delete(followerID, followeeID uint) error {
	return f.db.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).Delete(&domain.Follow{}).Error
}

func (f *followRepo) IsExist(followerID, followedID uint) (bool, error) {
	var follow domain.Follow
	tx := f.db.Model(&domain.Follow{}).Where("follower_id = ? AND followee_id = ?",
		followerID, followedID).Limit(1).Find(&follow)
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}

func (f *followRepo) ListFollowing(followerID uint, limit, offset int) ([]*domain.Follow, error) {
	var follows []*domain.Follow
	query := f.db.Where("follower_id = ?", followerID).Order("created_at desc")
	err := query.Limit(limit).Offset(offset).Preload("Followee").Find(&follows).Error
	return follows, err
}

func (f *followRepo) ListFollowers(followedID uint, limit, offset int) ([]*domain.Follow, error) {
	var follows []*domain.Follow
	query := f.db.Where("followee_id = ?", followedID).Order("created_at desc")
	err := query.Limit(limit).Offset(offset).Preload("Follower").Find(&follows).Error
	return follows, err
}

func (f *followRepo) BatchCheckFollowing(followerID uint, followeeIDs []uint) (map[uint]bool, error) {
	res := make(map[uint]bool)
	if len(followeeIDs) == 0 {
		return res, nil
	}

	var follows []*domain.Follow
	err := f.db.Select("followee_id").
		Where("follower_id = ? AND followee_id IN ?", followerID, followeeIDs).
		Find(&follows).Error

	if err != nil {
		return nil, err
	}

	for _, follow := range follows {
		res[follow.FolloweeID] = true
	}
	return res, nil
}

func (f *followRepo) CountFollowers(followedID uint) (int64, error) {
	var count int64
	err := f.db.Model(&domain.Follow{}).Where("followee_id = ?", followedID).Count(&count).Error
	return count, err
}

func (f *followRepo) CountFollowing(followerID uint) (int64, error) {
	var count int64
	err := f.db.Model(&domain.Follow{}).Where("follower_id = ?", followerID).Count(&count).Error
	return count, err
}
