package repository

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/geospackle/go-words-api/src/models"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type SearchRepository interface {
	Search(ctx context.Context, index []string, query string) (*models.SearchResult, error)
	Insert(ctx context.Context, index string, document models.Document) error
}

type OpenSearchClient struct {
	Client *opensearch.Client
}

func NewOpenSearchClient(username string, password string, host string) (*opensearch.Client, error) {
	client, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: []string{host},
		Username:  username,
		Password:  password,
	})

	return client, err
}

func (c *OpenSearchClient) Insert(ctx context.Context, index string, document models.Document) error {
	insertBody := strings.NewReader(document.Content)

	// auto id
	req := opensearchapi.IndexRequest{
		Index: index,
		Body:  insertBody,
	}

	_, err := req.Do(context.Background(), c.Client)

	return err
}

func (c *OpenSearchClient) Search(ctx context.Context, index []string, query string) (*models.SearchResult, error) {

	searchBody := strings.NewReader(query)
	search := opensearchapi.SearchRequest{
		Index: index,
		Body:  searchBody,
	}

	res, err := search.Do(context.Background(), c.Client)
	if err != nil {
		return nil, err
	}

	bodyBytes, _ := io.ReadAll(res.Body)
	var result models.SearchResult
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
