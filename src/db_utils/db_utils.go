package db_utils

import (
	"context"
	"strings"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func CreateIndex(client *opensearch.Client, index string) error {
	settings := strings.NewReader(`{
     "settings": {
       "index": {
            "number_of_shards": 1,
            "number_of_replicas": 0
            }
          }
     }`)

	// fielddata: true enables aggreagate query
	// with unknown memory implications
	mappings := strings.NewReader(`{
	"properties": {
	  "word": {
		"type": "text",
          "fielddata": true,
		  "fields": {
			"raw": {
			  "type": "keyword"
			  }
		  }
		}
	  }
     }`)

	var createReq = opensearchapi.IndicesCreateRequest{
		Index: index,
		Body:  settings,
	}

	_, err := createReq.Do(context.Background(), client)

	if err != nil {
		return err
	}

	var mappingReq = opensearchapi.IndicesPutMappingRequest{
		Index: []string{index},
		Body:  mappings,
	}

	_, err = mappingReq.Do(context.Background(), client)

	if err != nil {
		return err
	}

	return nil
}
