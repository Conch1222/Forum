package service

import (
	"Forum/internal/domain"
	"Forum/internal/repository"
	"context"
	"strings"
	"time"
)

type NotificationService interface {
	List(ctx context.Context, userID uint, unreadOnly bool, page, pageSize int) ([]domain.Notification, int64, error)
	UnreadCount(ctx context.Context, userID uint) (int64, error)
	MarkRead(ctx context.Context, userID uint, notificationID uint) error
	MarkAllRead(ctx context.Context, userID uint) (int64, error)

	// write notification
	Notify(ctx context.Context, userID uint, typ, entityType string, entityID uint, content string) error
}

type notificationServiceImpl struct {
	notificationRepo repository.NotificationRepo
}

func NewNotificationService(notificationRepo repository.NotificationRepo) NotificationService {
	return &notificationServiceImpl{notificationRepo: notificationRepo}
}

func (n *notificationServiceImpl) List(ctx context.Context, userID uint, unreadOnly bool, page, pageSize int) ([]domain.Notification, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	return n.notificationRepo.ListByUser(userID, unreadOnly, pageSize, offset)
}

func (n *notificationServiceImpl) UnreadCount(ctx context.Context, userID uint) (int64, error) {
	return n.notificationRepo.UnreadCount(userID)
}

func (n *notificationServiceImpl) MarkRead(ctx context.Context, userID uint, notificationID uint) error {
	return n.notificationRepo.MarkRead(userID, notificationID, time.Now())
}

func (n *notificationServiceImpl) MarkAllRead(ctx context.Context, userID uint) (int64, error) {
	return n.notificationRepo.MarkAllRead(userID, time.Now())
}

func (n *notificationServiceImpl) Notify(ctx context.Context, userID uint, typ, entityType string, entityID uint, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	notif := &domain.Notification{
		UserID:     userID,
		Type:       typ,
		EntityType: entityType,
		EntityID:   entityID,
		Content:    content,
	}
	return n.notificationRepo.Create(notif)
}
