package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/andriisoldatenko/go-news/api/internal/handlers"
	"github.com/andriisoldatenko/go-news/storage"
)

func main() {
	store, err := storage.NewBoltStore("articles.db", true)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	defer store.Close()
	h := handlers.New(store)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	fmt.Println("Starting API server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
