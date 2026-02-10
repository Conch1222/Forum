package repository

import (
	"Forum/internal/domain"
	"time"

	"gorm.io/gorm"
)

type NotificationRepo interface {
	Create(n *domain.Notification) error
	ListByUser(userID uint, unreadOnly bool, limit, offset int) ([]domain.Notification, int64, error)
	UnreadCount(userID uint) (int64, error)
	MarkRead(userID uint, notificationID uint, readAt time.Time) error
	MarkAllRead(userID uint, readAt time.Time) (int64, error) // return data count that affected
}

type notificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) NotificationRepo {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) Create(n *domain.Notification) error {
	return r.db.Create(n).Error
}

func (r *notificationRepo) ListByUser(userID uint, unreadOnly bool, limit, offset int) ([]domain.Notification, int64, error) {
	var notifications []domain.Notification
	var totalCount int64

	query := r.db.Model(&domain.Notification{}).
		Where("user_id = ?", userID)

	if unreadOnly {
		query = query.Where("read_at IS NULL")
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error

	return notifications, totalCount, err
}

func (r *notificationRepo) UnreadCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Notification{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *notificationRepo) MarkRead(userID uint, notificationID uint, readAt time.Time) error {
	return r.db.Model(&domain.Notification{}).
		Where("id = ? AND user_id = ? AND read_at IS NULL", notificationID, userID).
		Updates(map[string]interface{}{"read_at": readAt}).Error
}

func (r *notificationRepo) MarkAllRead(userID uint, readAt time.Time) (int64, error) {
	tx := r.db.Model(&domain.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Updates(map[string]interface{}{"read_at": readAt})

	return tx.RowsAffected, tx.Error
}
