package repository

import (
	"Forum/internal/domain"

	"gorm.io/gorm"
)

type PostRepo interface {
	Create(post *domain.Post) error
	FindByID(id uint) (*domain.Post, error)
	Update(post *domain.Post) error
	Delete(id uint) error
	List(limit, offset int) ([]*domain.Post, int64, error) // return list of posts and total count
	IncrementViewCount(id uint) error
	IncrementLikeCount(id uint, delta int) error
}

type postRepo struct {
	db *gorm.DB
}

func NewPostRepo(db *gorm.DB) PostRepo {
	return &postRepo{db: db}
}

func (p *postRepo) Create(post *domain.Post) error {
	return p.db.Create(post).Error
}

func (p *postRepo) FindByID(id uint) (*domain.Post, error) {
	var post domain.Post
	err := p.db.First(&post, id).Error
	return &post, err
}

func (p *postRepo) Update(post *domain.Post) error {
	return p.db.Save(post).Error
}

func (p *postRepo) Delete(id uint) error {
	return p.db.Delete(&domain.Post{}, id).Error // soft delete
}

func (p *postRepo) List(limit, offset int) ([]*domain.Post, int64, error) {
	var posts []*domain.Post
	var totalCount int64

	// only show published posts
	query := p.db.Model(&domain.Post{}).Where("status=?", "published")

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// pagination
	err := query.Preload("User").
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	return posts, totalCount, err
}

func (p *postRepo) IncrementViewCount(id uint) error {
	query := p.db.Model(&domain.Post{}).Where("id=?", id)

	return query.UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (p *postRepo) IncrementLikeCount(id uint, delta int) error {
	return p.db.Model(&domain.Post{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("GREATEST(like_count + ?, 0)", delta)).Error
}
