package main

import (
	"context"
	"log"
	"net/http"

	"github.com/andriisoldatenko/go-news/api/internal/handlers"
	"github.com/andriisoldatenko/go-news/storage"
)

func main() {
	baseStore, err := storage.NewBoltStore("articles.db", true)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	defer baseStore.Close()

	store, err := storage.NewSearchStoreWithDefaults(context.Background(), baseStore)
	if err != nil {
		log.Fatal(err)
	}

	articleHandlers := handlers.NewArticleHandlers(store)
	http.HandleFunc("/articles", articleHandlers.HandleArticles)

	searchHandlers := handlers.NewSearchHandlers(store)
	http.HandleFunc("/search", searchHandlers.HandleSearch)

	log.Println("API server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
