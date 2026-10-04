package main

import (
	"context"
	"fmt"
	"log"

	"github.com/andriisoldatenko/go-news/domain"
	"github.com/andriisoldatenko/go-news/storage"
	"github.com/andriisoldatenko/go-news/worker/internal/reader"
)

var feeds = []string{
	"https://rss.nytimes.com/services/xml/rss/nyt/World.xml",
	"https://feeds.bbci.co.uk/news/rss.xml",
}

func fetchAndStore(ctx context.Context, fetcher domain.Fetcher, store domain.Storage, urls []string) {
	for _, url := range urls {
		feed, err := fetcher.FetchFeed(ctx, url)
		if err != nil {
			log.Printf("Error fetching %s: %v", url, err)
			continue
		}

		if err := store.AddArticles(ctx, feed.Articles); err != nil {
			log.Printf("Error storing articles from %s: %v", url, err)
			continue
		}

		fmt.Printf("Fetched and stored %d articles from %s\n",
			len(feed.Articles), feed.Title)
	}
}

func main() {
	baseStore, err := storage.NewBoltStore("articles.db", false)

	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	defer baseStore.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := storage.NewSearchStoreWithDefaults(ctx, baseStore)
	if err != nil {
		log.Printf("Warning: Could not enable search: %v", err)
		log.Println("Continuing with basic storage only")
	}

	fetcher := reader.NewRSSReader()

	//sigChan := make(chan os.Signal, 1)
	//
	//signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	//
	//go func() {
	//	<-sigChan
	//	cancel()
	//	fmt.Println("\nShutdown signal received, stopping worker...")
	//}()
	//
	//fmt.Println("Worker started, fetching feeds every 5 minutes...")

	//ticker := time.NewTicker(5 * time.Minute)

	//defer ticker.Stop()

	fetchAndStore(ctx, fetcher, store, feeds)

	//for {
	//	select {
	//	case <-ticker.C:
	//		fetchAndStore(ctx, fetcher, store, feeds)
	//	case <-ctx.Done():
	//		fmt.Println("Worker stopped gracefully")
	//		return
	//	}
	//}
}
