package repository

import (
	"Forum/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type SearchRepo interface {
	SearchPosts(ctx context.Context, q string, limit, offset int) ([]domain.PostResponse, int64, error)
}

type searchRepo struct {
	es *opensearchapi.Client
}

func NewSearchRepo(es *opensearchapi.Client) SearchRepo {
	return &searchRepo{es: es}
}

// SearchPosts use elastic search
func (s *searchRepo) SearchPosts(ctx context.Context, q string, limit, offset int) ([]domain.PostResponse, int64, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []domain.PostResponse{}, 0, nil
	}

	query := buildPostSearchQuery(q, limit, offset)

	req := opensearchapi.SearchReq{
		Indices: []string{"posts_v1"},
		Body:    strings.NewReader(query),
	}

	httpReq, err := req.GetRequest()
	if err != nil {
		return nil, 0, err
	}
	httpReq = httpReq.WithContext(ctx)

	res, err := s.es.Client.Perform(httpReq)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return nil, 0, fmt.Errorf("search failed: status=%d body=%s", res.StatusCode, string(body))
	}

	var response domain.EsPostSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, 0, err
	}

	posts := make([]domain.PostResponse, 0, len(response.Hits.Hits))
	for _, hit := range response.Hits.Hits {
		userID, _ := strconv.ParseUint(hit.Source.UserID, 10, 64)
		postID, _ := strconv.ParseUint(hit.Source.ID, 10, 64)

		posts = append(posts, domain.PostResponse{
			ID:        uint(postID),
			UserID:    uint(userID),
			Title:     hit.Source.Title,
			Content:   hit.Source.Content,
			LikeCount: hit.Source.LikeCount,
			CreatedAt: hit.Source.CreatedAt,
			UpdatedAt: hit.Source.UpdatedAt,
		})
	}

	return posts, response.Hits.Total.Value, nil
}

func buildPostSearchQuery(q string, limit, offset int) string {
	query := fmt.Sprintf(`{
	  "track_total_hits": true,
	  "from": %d,
	  "size": %d,
	  "sort": [
		{ "_score": { "order": "desc" } },
		{ "created_at": { "order": "desc" } }
	  ],
	  "query": {
		"bool": {
		  "must": [
			{
			  "multi_match": {
				"query": %q,
				"fields": ["title^2", "content"]
			  }
			}
		  ],
		  "filter": [
			{
			  "term": {
				"status": "published"
			  }
			}
		  ]
		}
	  }
	}`, offset, limit, q)

	return query
}
