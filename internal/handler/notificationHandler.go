package handler

import (
	"Forum/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService service.NotificationService
}

func NewNotificationHandler(notificationHandler service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationHandler}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	unreadOnly := c.DefaultQuery("unread_only", "0") == "1"

	notifs, totalCount, err := h.notificationService.List(c.Request.Context(), userID, unreadOnly, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifs,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       totalCount,
			"total_pages": (totalCount + int64(pageSize) - 1) / int64(pageSize), // Round up
		},
	})
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := c.GetUint("user_id")

	cnt, err := h.notificationService.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": cnt})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := c.GetUint("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.notificationService.MarkRead(c.Request.Context(), userID, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := c.GetUint("user_id")

	affected, err := h.notificationService.MarkAllRead(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"affected": affected})
}
