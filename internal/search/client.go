package search

import (
	"Forum/internal/config"

	"github.com/elastic/go-elasticsearch/v8"
)

func NewESClient(cfg *config.Config) (*elasticsearch.Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.OpenSearchURL},
		Username:  cfg.OpenSearchUser,
		Password:  cfg.OpenSearchPassword,
	}
	return elasticsearch.NewClient(esCfg)
}
