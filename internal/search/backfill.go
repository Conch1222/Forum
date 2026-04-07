package search

import (
	"Forum/internal/domain"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
	"gorm.io/gorm"
)

type BackfillRunner struct {
	Db *gorm.DB
	Bi opensearchutil.BulkIndexer
}

// Run migrate postgres posts table to elastic search
func (r *BackfillRunner) Run(ctx context.Context) error {
	const batchSize = 200
	var lastID uint

	for {
		var posts []domain.Post
		err := r.Db.WithContext(ctx).
			Where("deleted_at IS NULL").
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(batchSize).
			Find(&posts).Error
		if err != nil {
			return err
		}

		if len(posts) == 0 {
			break
		}

		for _, post := range posts {
			doc := domain.NewPostSearchDoc(post)

			data, err := json.Marshal(doc)
			if err != nil {
				return err
			}

			err = r.Bi.Add(ctx, opensearchutil.BulkIndexerItem{
				Action:     "index",
				DocumentID: strconv.FormatUint(uint64(post.ID), 10),
				Body:       bytes.NewReader(data),
				OnFailure: func(ctx context.Context, item opensearchutil.BulkIndexerItem, res opensearchapi.BulkRespItem, err error) {
					if err != nil {
						log.Printf("bulk indexer error: %v", err)
						return
					}
					log.Printf("bulk item failed: type=%s reason=%s", res.Error.Type, res.Error.Reason)
				},
			})
			if err != nil {
				return err
			}
		}

		lastID = posts[len(posts)-1].ID
		log.Printf("indexed through post_id=%d", lastID)
	}

	return r.Bi.Close(ctx)
}

func EnsurePostsIndex(es *opensearchapi.Client) error {
	const indexName = "posts_v1"
	const mappingFile = "internal/search/mappings/posts_v1.json"

	existsReq := opensearchapi.IndicesExistsReq{
		Indices: []string{indexName},
	}
	httpReq, err := existsReq.GetRequest()
	if err != nil {
		return fmt.Errorf("failed to build exists request: %w", err)
	}

	existsResp, err := es.Client.Perform(httpReq)
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}
	defer existsResp.Body.Close()

	if existsResp.StatusCode == http.StatusOK {
		return nil
	}

	if existsResp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(existsResp.Body)
		return fmt.Errorf("unexpected exists response: status=%d, body=%s", existsResp.StatusCode, string(body))
	}

	data, err := os.ReadFile(mappingFile)
	if err != nil {
		return fmt.Errorf("failed to read mapping file: %w", err)
	}

	createReq := opensearchapi.IndicesCreateReq{
		Index: indexName,
		Body:  strings.NewReader(string(data)),
	}

	createHTTPReq, err := createReq.GetRequest()
	if err != nil {
		return fmt.Errorf("failed to build create index request: %w", err)
	}

	createResp, err := es.Client.Perform(createHTTPReq)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode >= 300 {
		body, _ := io.ReadAll(createResp.Body)
		return fmt.Errorf("create index failed: status=%d, body=%s", createResp.StatusCode, string(body))
	}
	return nil
}

func NewPostBulkIndexer(es *opensearchapi.Client) (opensearchutil.BulkIndexer, error) {
	return opensearchutil.NewBulkIndexer(opensearchutil.BulkIndexerConfig{
		Client:        es,
		Index:         "posts_v1",
		NumWorkers:    2,
		FlushBytes:    5e+6,
		FlushInterval: 5 * time.Second,
	})
}
