package handler

import (
	"Forum/internal/domain"
	"Forum/internal/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type LikeHandler struct {
	likeService service.LikeService
}

func NewLikeHandler(likeService service.LikeService) *LikeHandler {
	return &LikeHandler{likeService: likeService}
}

func (h *LikeHandler) Toggle(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req domain.ToggleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	}

	res, err := h.likeService.Toggle(c.Request.Context(), userID, req.TargetID, req.TargetType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *LikeHandler) GetStatus(c *gin.Context) {
	userID := c.GetUint("user_id")

	targetType := strings.ToLower(strings.TrimSpace(c.Param("type")))
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	targetID := uint(id)

	res, err := h.likeService.GetStatus(c.Request.Context(), userID, targetID, targetType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, res)
}
