package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/geospackle/go-words-api/src/repository"
)

type Payload struct {
	Word string `json:"word"`
}

type Response struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
}

func ValidateValue(value string) bool {
	regex := regexp.MustCompile(`^[a-zA-Z]+$`)
	return regex.MatchString(value)
}

func PostHandler(w http.ResponseWriter, r *http.Request, index string, repo repository.OpenSearchRepository) {
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
		http.Error(w, fmt.Sprintf("Invalid input payload. Needs to be format `{\"word\":\"<singleWord>\"}`."), http.StatusBadRequest)
		return
	}

	document := `{"word":"` + payload.Word + `"}`

	err = repo.Insert(context.Background(), index, document)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting value: %s", err), http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Successfully inserted value: %s", payload.Word)
}

func GetHandler(w http.ResponseWriter, r *http.Request, indexes []string, repo repository.OpenSearchRepository) {
	value := r.URL.Query().Get("prefix")

	if !ValidateValue(value) {
		http.Error(w, "Needs query parameter 'word'", http.StatusBadRequest)
		return
	}

	//https://opensearch.org/docs/latest/aggregations/bucket/terms/
	//https://opensearch.org/docs/latest/aggregations/
	query := `{
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
	}`

	res, err := repo.Search(context.Background(), indexes, query)

	if err != nil {
		http.Error(w, fmt.Sprintf("Query can not be processed: %s", err), http.StatusBadGateway)
		return
	}

	var msg string
	dataOut := res.Aggregations.MaxDistinctCounts

	if dataOut.Value == 0 {
		msg = "No results retrieved"
	} else {
		msg = "Results retrieved"
	}

	json.NewEncoder(w).Encode(Response{
		StatusCode: http.StatusOK,
		Message:    msg,
		Data:       dataOut,
	})
}
