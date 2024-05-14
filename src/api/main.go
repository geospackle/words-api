package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

// Define a struct to hold the request data
type Payload struct {
	Word string `json:"word"`
}

type Response struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
}
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

type OpenSearchResponse struct {
	Aggregations Aggregations
}

var client, _ = opensearch.NewClient(opensearch.Config{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
	Addresses: []string{"https://localhost:9200"},
	Username:  "admin",
	Password:  "badDamin$33",
})

func CreateIndex() {
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
		Index: "index-fielddata6",
		Body:  settings,
	}

	_, err := createReq.Do(context.Background(), client)

	if err != nil {
		fmt.Println(err)
	}

	var mappingReq = opensearchapi.IndicesPutMappingRequest{
		Index: []string{"index-fielddata6"},
		Body:  mappings,
	}

	_, err = mappingReq.Do(context.Background(), client)

	if err != nil {
		fmt.Println(err)
	}
}

// ValidateValue function checks if the provided string matches the regex
func ValidateValue(value string) bool {
	regex := regexp.MustCompile(`^[a-zA-Z]+$`)
	return regex.MatchString(value)
}

func HelloHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Hello, world!\n")
}

// CASE INSENSITIVE!
func PostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload Payload
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)

	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if !ValidateValue(payload.Word) {
		http.Error(w, fmt.Sprintf("Invalid input: %s. Needs to be a single word.", payload.Word), http.StatusBadRequest)
		return
	}

	document := strings.NewReader(`{"word":"` + payload.Word + `"}`)

	// auto id
	req := opensearchapi.IndexRequest{
		Index: "index-fielddata6",
		Body:  document,
	}

	res, err := req.Do(context.Background(), client)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting value: %s", err), http.StatusBadGateway)
	}

	fmt.Fprintf(w, "Successfully inserted value: %s", res)
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	value := r.URL.Query().Get("word")

	if !ValidateValue(value) {
		http.Error(w, "Needs query parameter 'word'", http.StatusBadRequest)
		return
	}

	//https://opensearch.org/docs/latest/aggregations/bucket/terms/
	//https://opensearch.org/docs/latest/aggregations/
	content := strings.NewReader(`{
	  "size": 0,
	  "query": {
		"bool": {
		  "filter": {
			"prefix": {
			  "word": {
				"value": "` + value + `",
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
	}`)

	search := opensearchapi.SearchRequest{
		Index: []string{"index-fielddata6"},
		Body:  content,
	}

	res, err := search.Do(context.Background(), client)

	if err != nil {
		http.Error(w, fmt.Sprintf("Query can not be processed: %s", err), http.StatusBadGateway)
		return
	}

	bodyBytes, _ := io.ReadAll(res.Body)

	var osRes OpenSearchResponse
	err = json.Unmarshal(bodyBytes, &osRes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query can not be processed: %s", err), http.StatusBadGateway)
		return
	}

	json.NewEncoder(w).Encode(Response{
		StatusCode: http.StatusOK,
		Message:    "Data retrieved successfully",
		Data:       osRes.Aggregations.MaxDistinctCounts,
	})
}
