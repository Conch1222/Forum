package service

import (
	"Forum/internal/domain"
	"Forum/internal/metrics"
	"Forum/internal/repository"
	"Forum/internal/search"
	"context"
	"errors"

	"gorm.io/gorm"
)

type PostService interface {
	Create(ctx context.Context, userID uint, req *domain.CreatePostRequest) (*domain.PostResponse, error)
	GetByID(ctx context.Context, postID uint) (*domain.PostResponse, error)
	Update(ctx context.Context, userID, postID uint, req *domain.UpdatePostRequest) error
	Delete(ctx context.Context, userID, postID uint) error
	List(ctx context.Context, page, pageSize int) ([]*domain.PostResponse, int64, error)
}

type postServiceImpl struct {
	postRepo    repository.PostRepo
	userRepo    repository.UserRepo
	postIndexer search.PostIndexer
}

func NewPostService(postRepo repository.PostRepo, userRepo repository.UserRepo, postIndexer search.PostIndexer) PostService {
	return &postServiceImpl{postRepo: postRepo, userRepo: userRepo, postIndexer: postIndexer}
}

func (p *postServiceImpl) Create(ctx context.Context, userID uint, req *domain.CreatePostRequest) (*domain.PostResponse, error) {
	// default status
	if req.Status == "" {
		req.Status = "published"
	}

	post := &domain.Post{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
		Status:  req.Status,
	}

	if err := p.postRepo.Create(post); err != nil {
		return nil, err
	}

	// es index post
	if err := p.postIndexer.IndexPost(ctx, post); err != nil {
		return nil, err
	}

	user, _ := p.userRepo.FindByID(userID)

	metrics.PostsCreatedTotal.Inc() // for prometheus

	return &domain.PostResponse{
		ID:        post.ID,
		UserID:    user.ID,
		UserName:  user.UserName,
		Title:     post.Title,
		Content:   post.Content,
		Status:    post.Status,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}, nil
}

func (p *postServiceImpl) GetByID(ctx context.Context, postID uint) (*domain.PostResponse, error) {
	post, err := p.postRepo.FindByID(postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}

	// increases view count
	err = p.postRepo.IncrementViewCount(postID)
	if err != nil {
		return nil, err
	}

	return &domain.PostResponse{
		ID:        post.ID,
		UserID:    post.UserID,
		UserName:  post.User.UserName,
		Title:     post.Title,
		Content:   post.Content,
		Status:    post.Status,
		ViewCount: post.ViewCount + 1, // increment view count
		LikeCount: post.LikeCount,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}, nil
}

func (p *postServiceImpl) Update(ctx context.Context, userID, postID uint, req *domain.UpdatePostRequest) error {
	post, err := p.postRepo.FindByID(postID)
	if err != nil {
		return errors.New("post not found")
	}

	// check permission, only allow user to update their own post
	if post.UserID != userID {
		return errors.New("permission denied")
	}

	// update column
	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}
	if req.Status != "" {
		post.Status = req.Status
	}
	err = p.postRepo.Update(post)
	if err != nil {
		return err
	}

	// es index post
	if err := p.postIndexer.IndexPost(ctx, post); err != nil {
		return err
	}

	return nil
}

func (p *postServiceImpl) Delete(ctx context.Context, userID, postID uint) error {
	post, err := p.postRepo.FindByID(postID)
	if err != nil {
		return errors.New("post not found")
	}

	// check permission, only allow user to update their own post
	if post.UserID != userID {
		return errors.New("permission denied")
	}
	err = p.postRepo.Delete(postID)
	if err != nil {
		return err
	}

	// es delete post
	if err := p.postIndexer.DeletePost(ctx, postID); err != nil {
		return err
	}

	return nil
}

func (p *postServiceImpl) List(ctx context.Context, page, pageSize int) ([]*domain.PostResponse, int64, error) {
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
	posts, totalCount, err := p.postRepo.List(pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	var res []*domain.PostResponse
	for _, post := range posts {
		res = append(res, &domain.PostResponse{
			ID:        post.ID,
			UserID:    post.UserID,
			UserName:  post.User.UserName,
			Title:     post.Title,
			Content:   post.Content,
			Status:    post.Status,
			ViewCount: post.ViewCount,
			LikeCount: post.LikeCount,
			CreatedAt: post.CreatedAt,
			UpdatedAt: post.UpdatedAt,
		})
	}
	return res, totalCount, nil
}
