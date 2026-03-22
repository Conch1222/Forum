package search

import (
	"Forum/internal/domain"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8"
)

type PostIndexer interface {
	IndexPost(ctx context.Context, post *domain.Post) error
	DeletePost(ctx context.Context, postId uint) error
}

type PostIndexerImpl struct {
	es *elasticsearch.Client
}

func NewPostIndexer(es *elasticsearch.Client) PostIndexer {
	return &PostIndexerImpl{es: es}
}

func (p *PostIndexerImpl) IndexPost(ctx context.Context, post *domain.Post) error {
	doc := map[string]any{
		"id":         strconv.FormatUint(uint64(post.ID), 10),
		"user_id":    strconv.FormatUint(uint64(post.UserID), 10),
		"title":      post.Title,
		"content":    post.Content,
		"status":     post.Status,
		"view_count": post.ViewCount,
		"like_count": post.LikeCount,
		"created_at": post.CreatedAt,
		"updated_at": post.UpdatedAt,
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	res, err := p.es.Index("posts_v1",
		bytes.NewBuffer(body),
		p.es.Index.WithDocumentID(strconv.FormatUint(uint64(post.ID), 10)),
		p.es.Index.WithContext(ctx),
	)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index post to es failed: %s", res.String())
	}

	return nil
}

func (p *PostIndexerImpl) DeletePost(ctx context.Context, postID uint) error {
	res, err := p.es.Delete(
		"posts_v1",
		strconv.FormatUint(uint64(postID), 10),
		p.es.Delete.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("delete post from es failed: %s", res.String())
	}

	return nil
}
