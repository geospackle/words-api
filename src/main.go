package main

import (
	"fmt"
	"net/http"

	"github.com/geospackle/go-words-api/src/api"
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {

		fmt.Fprintf(w, "API is healthy!")
	})
	api.CreateIndex()

	router.HandleFunc("/hello", api.HelloHandler)

	router.HandleFunc("/post", api.PostHandler)
	router.HandleFunc("/get", api.GetHandler)

	fmt.Println("Starting API server on port 8080")
	http.ListenAndServe(":8081", router)
}
