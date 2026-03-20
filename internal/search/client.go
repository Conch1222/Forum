package search

import "github.com/elastic/go-elasticsearch/v8"

func NewESClient() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://elasticsearch:9200"},
	}
	return elasticsearch.NewClient(cfg)
}
