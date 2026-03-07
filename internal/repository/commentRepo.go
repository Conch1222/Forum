package repository

import (
	"Forum/internal/domain"

	"gorm.io/gorm"
)

type CommentRepo interface {
	WithTx(tx *gorm.DB) CommentRepo
	Create(comment *domain.Comment) error
	FindByID(id uint) (*domain.Comment, error)
	ListByPostID(postID uint, limit, offset int) ([]domain.Comment, error)
	Delete(id uint) error
	IncrementLikeCount(id uint, delta int) error
}

type commentRepo struct {
	db *gorm.DB
}

func NewCommentRepo(db *gorm.DB) CommentRepo {
	return &commentRepo{db: db}
}

func (r *commentRepo) WithTx(tx *gorm.DB) CommentRepo {
	return &commentRepo{db: tx}
}

func (c *commentRepo) Create(comment *domain.Comment) error {
	return c.db.Create(comment).Error
}

func (c *commentRepo) FindByID(id uint) (*domain.Comment, error) {
	var comment domain.Comment
	err := c.db.First(&comment, id).Error
	return &comment, err
}

func (c *commentRepo) ListByPostID(postID uint, limit, offset int) ([]domain.Comment, error) {
	var comments []domain.Comment

	query := c.db.Model(&domain.Comment{}).Where("post_id=?", postID)
	err := query.Preload("User").
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&comments).Error

	return comments, err
}

func (c *commentRepo) Delete(id uint) error {
	return c.db.Delete(&domain.Comment{}, id).Error
}

func (c *commentRepo) IncrementLikeCount(id uint, delta int) error {
	return c.db.Model(&domain.Comment{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("GREATEST(like_count + ?, 0)", delta)).Error
}
