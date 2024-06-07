package service

import (
	"context"
	"os"

	"github.com/geospackle/go-words-api/src/models"
	"github.com/geospackle/go-words-api/src/repository"
)

var INDEX = os.Getenv("INDEX")

type Service interface {
	InsertDocument(context.Context, models.Document) error
	SearchByPrefix(context.Context, string) (*models.SearchResult, error)
}

type service struct {
	searchRepo repository.SearchRepository
}

func NewService(searchRepo repository.SearchRepository) Service {
	return &service{
		searchRepo: searchRepo,
	}
}

func (s service) InsertDocument(ctx context.Context, document models.Document) error {
	err := s.searchRepo.Insert(ctx, INDEX, document)
	return err
}

func (s service) SearchByPrefix(ctx context.Context, prefix string) (*models.SearchResult, error) {
	//https://opensearch.org/docs/latest/aggregations/bucket/terms/
	//https://opensearch.org/docs/latest/aggregations/
	query := `{
	  "size": 0,
	  "query": {
		"bool": {
		  "filter": {
			"prefix": {
			  "word": {
				"value": "` + prefix + `",
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

	res, err := s.searchRepo.Search(ctx, []string{INDEX}, query)
	return res, err
}
