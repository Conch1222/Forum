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

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"gorm.io/gorm"
)

type BackfillRunner struct {
	Db *gorm.DB
	Bi esutil.BulkIndexer
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
			Find(&posts).Error
		if err != nil {
			return err
		}

		if len(posts) == 0 {
			break
		}

		for _, post := range posts {
			doc := NewPostSearchDoc(post)

			data, err := json.Marshal(doc)
			if err != nil {
				return err
			}

			err = r.Bi.Add(ctx, esutil.BulkIndexerItem{
				Action:     "index",
				DocumentID: strconv.FormatUint(uint64(post.ID), 10),
				Body:       bytes.NewReader(data),
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
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

func EnsurePostsIndex(es *elasticsearch.Client) error {
	const indexName = "posts_v1"
	const mappingFile = "internal/search/mappings/posts_v1.json"

	existsResp, err := es.Indices.Exists([]string{indexName})
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

	createResp, err := es.Indices.Create(indexName, es.Indices.Create.WithBody(strings.NewReader(string(data))))
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer createResp.Body.Close()

	if createResp.IsError() {
		body, _ := io.ReadAll(createResp.Body)
		return fmt.Errorf("create index failed: status=%d, body=%s", createResp.StatusCode, string(body))
	}
	return nil
}

func NewPostBulkIndexer(es *elasticsearch.Client) (esutil.BulkIndexer, error) {
	return esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:        es,
		Index:         "posts_v1",
		NumWorkers:    2,
		FlushBytes:    5e+6,
		FlushInterval: 5 * time.Second,
	})
}
