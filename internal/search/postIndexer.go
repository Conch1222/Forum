package search

import (
	"Forum/internal/domain"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
)

type PostIndexer interface {
	IndexPost(ctx context.Context, post *domain.Post) error
	DeletePost(ctx context.Context, postId uint) error
}

type PostIndexerImpl struct {
	es *opensearchapi.Client
}

func NewPostIndexer(es *opensearchapi.Client) PostIndexer {
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

	req := opensearchapi.IndexReq{
		Index:      "posts_v1",
		DocumentID: strconv.FormatUint(uint64(post.ID), 10),
		Body:       opensearchutil.NewJSONReader(doc),
	}

	httpReq, err := req.GetRequest()
	if err != nil {
		return err
	}
	httpReq = httpReq.WithContext(ctx)

	res, err := p.es.Client.Perform(httpReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("index post to opensearch failed: status=%d, body=%s", res.StatusCode, string(body))
	}

	return nil
}

func (p *PostIndexerImpl) DeletePost(ctx context.Context, postID uint) error {
	req := opensearchapi.DocumentDeleteReq{
		Index:      "posts_v1",
		DocumentID: strconv.FormatUint(uint64(postID), 10),
	}

	httpReq, err := req.GetRequest()
	if err != nil {
		return err
	}
	httpReq = httpReq.WithContext(ctx)

	res, err := p.es.Client.Perform(httpReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 && res.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete post from opensearch failed: status=%d body=%s", res.StatusCode, string(body))
	}

	return nil
}
