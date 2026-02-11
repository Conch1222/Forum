package service

import (
	"Forum/internal/domain"
	"Forum/internal/repository"
	"context"
	"errors"
)

type CommentService interface {
	Create(userID, postID uint, req *domain.CreateCommentRequest) (*domain.CommentResponse, error)
	ListByPostID(postID uint, page, pageSize int) ([]*domain.CommentResponse, error)
	Delete(userID, commentID uint) error
}

type commentServiceImpl struct {
	commentRepo         repository.CommentRepo
	postRepo            repository.PostRepo
	userRepo            repository.UserRepo
	notificationService NotificationService
}

func NewCommentService(commentRepo repository.CommentRepo, postRepo repository.PostRepo, userRepo repository.UserRepo, notifService NotificationService) CommentService {
	return &commentServiceImpl{commentRepo: commentRepo, postRepo: postRepo, userRepo: userRepo, notificationService: notifService}
}

func (c *commentServiceImpl) Create(userID, postID uint, req *domain.CreateCommentRequest) (*domain.CommentResponse, error) {
	// check if post exists
	post, err := c.postRepo.FindByID(postID)
	if err != nil {
		return nil, errors.New("post not found")
	}

	comment := &domain.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: req.Content,
	}

	if err := c.commentRepo.Create(comment); err != nil {
		return nil, err
	}

	// create notification
	if post.UserID != userID {
		_ = c.notificationService.Notify(context.Background(),
			post.UserID, // receive notif
			"comment_created",
			"post",
			post.ID,
			"someone commented in your post",
		)
	}

	user, _ := c.userRepo.FindByID(userID)
	return &domain.CommentResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		UserID:    comment.UserID,
		UserName:  user.UserName,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}, nil
}

func (c *commentServiceImpl) ListByPostID(postID uint, page, pageSize int) ([]*domain.CommentResponse, error) {
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
	comments, err := c.commentRepo.ListByPostID(postID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	var res []*domain.CommentResponse
	for _, comment := range comments {
		res = append(res, &domain.CommentResponse{
			ID:        comment.ID,
			PostID:    comment.PostID,
			UserID:    comment.UserID,
			UserName:  comment.User.UserName,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt,
		})
	}
	return res, nil
}

func (c *commentServiceImpl) Delete(userID, commentID uint) error {
	comment, err := c.commentRepo.FindByID(commentID)
	if err != nil {
		return errors.New("comment not found")
	}

	// check permission
	if comment.UserID != userID {
		return errors.New("permission denied")
	}

	return c.commentRepo.Delete(commentID)
}
