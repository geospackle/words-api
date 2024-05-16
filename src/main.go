package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	controller "github.com/geospackle/go-words-api/src/api"
	"github.com/geospackle/go-words-api/src/db_utils"
	"github.com/geospackle/go-words-api/src/repository"
	"github.com/opensearch-project/opensearch-go"
)

func main() {

	//os.Setenv("UNAME", "admin")
	//os.Setenv("PWORD", "badDamin$33")
	//os.Setenv("OPENSEARCH_HOST", "https://localhost:9200")
	//os.Setenv("INDEX", "test-index")
	var UNAME = os.Getenv("UNAME")
	var PWORD = os.Getenv("PWORD")
	var OS_HOST = os.Getenv("OPENSEARCH_HOST")
	var INDEX = os.Getenv("INDEX")

	maxRetries := 30
	retryDelay := 5 * time.Second
	var client *opensearch.Client
	var err error

	client, err = db_utils.CreateOpenSearchClient(UNAME, PWORD, OS_HOST)

	if err != nil {
		panic(fmt.Sprintf("Could not create opensearch client: %v", err))
	}

	for i := 1; i < maxRetries; i++ {
		db_utils.CreateIndex(client, INDEX)
		if err == nil {
			fmt.Println("Successfully created index!")
			break
		}

		fmt.Printf("Error creating index (attempt %d). Retrying: %v\n", i, err)
		time.Sleep(retryDelay)
	}

	if client == nil {
		panic("Not able to connect to open search")
	}

	repo := &repository.OpenSearchClient{Client: client}

	router := http.NewServeMux()
	router.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		controller.PostHandler(w, r, INDEX, repo)
	})
	router.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		controller.GetHandler(w, r, []string{INDEX}, repo)
	})
	fmt.Println("Starting API server on port 8080")
	http.ListenAndServe(":8080", router)
}
