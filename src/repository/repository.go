package repository

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type Bucket struct {
	DocCount int    `json:"doc_count"`
	Key      string `json:"key"`
}

type Aggregations struct {
	DistinctValueCount struct {
		Buckets                 []Bucket `json:"buckets"`
		DocCountErrorUpperBound int      `json:"doc_count_error_upper_bound"`
		SumOtherDocCount        int      `json:"sum_other_doc_count"`
	} `json:"distinct_value_count"`
	MaxDistinctCounts struct {
		Keys  []string `json:"keys"`
		Value float32  `json:"value"`
	} `json:"max_distinct_counts"`
}

type SearchResult struct {
	Hits struct {
		Hits []Hit `json:"hits"`
	} `json:"hits"`
	Aggregations `json:"aggregations"`
}

type Hit struct {
	Source map[string]interface{} `json:"_source"`
}
type OpenSearchRepository interface {
	Search(ctx context.Context, index []string, query string) (*SearchResult, error)
	Insert(ctx context.Context, index string, document string) error
}

type OpenSearchClient struct {
	Client *opensearch.Client
}

func (c *OpenSearchClient) Insert(ctx context.Context, index string, document string) error {
	insertBody := strings.NewReader(document)

	// auto id
	req := opensearchapi.IndexRequest{
		Index: index,
		Body:  insertBody,
	}

	_, err := req.Do(context.Background(), c.Client)

	return err
}

func (c *OpenSearchClient) Search(ctx context.Context, index []string, searchTerm string) (*SearchResult, error) {

	//https://opensearch.org/docs/latest/aggregations/bucket/terms/
	//https://opensearch.org/docs/latest/aggregations/
	query := `{
	  "size": 0,
	  "query": {
		"bool": {
		  "filter": {
			"prefix": {
			  "word": {
				"value": "` + searchTerm + `",
				"case_insensitive": true
			  }
			}
		  }
		}
	  },
	  "aggs": {
		"distinct_value_count": {
		  "terms": {
			"field": "word.raw",
			"size": 10
		  }
		},
	    "max_distinct_counts": {
          "max_bucket": {
            "buckets_path": "distinct_value_count>_count"
          }
        }
	  }
	}`

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
	var result SearchResult
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
