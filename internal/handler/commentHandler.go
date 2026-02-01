package handler

import (
	"Forum/internal/domain"
	"Forum/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	commentService service.CommentService
}

func NewCommentHandler(commentService service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

func (h *CommentHandler) CreateComment(c *gin.Context) {
	userID := c.GetUint("user_id")
	postID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req domain.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	}

	comment, err := h.commentService.Create(userID, uint(postID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *CommentHandler) ListComments(c *gin.Context) {
	postID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	comments, err := h.commentService.ListByPostID(uint(postID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (h *CommentHandler) DeleteComment(c *gin.Context) {
	userID := c.GetUint("user_id")
	commentID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.commentService.Delete(userID, uint(commentID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment deleted successfully"})
}
