package search

import (
	"Forum/internal/config"

	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

func NewESClient(cfg *config.Config) (*opensearchapi.Client, error) {
	return opensearchapi.NewClient(opensearchapi.Config{
		Client: opensearch.Config{
			Addresses: []string{cfg.OpenSearchURL},
			Username:  cfg.OpenSearchUser,
			Password:  cfg.OpenSearchPassword,
		},
	})
}
