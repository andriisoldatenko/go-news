package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

		if err := store.AddArticles(feed.Articles); err != nil {
			log.Printf("Error storing articles from %s: %v", url, err)
			continue
		}

		fmt.Printf("Fetched and stored %d articles from %s\n",
			len(feed.Articles), feed.Title)
	}
}

func main() {
	store, err := storage.NewBoltStore("articles.db", false)

	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	defer store.Close()

	fetcher := reader.NewRSSReader()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
		fmt.Println("\nShutdown signal received, stopping worker...")
	}()

	fmt.Println("Worker started, fetching feeds every 5 minutes...")

	ticker := time.NewTicker(5 * time.Minute)

	defer ticker.Stop()

	fetchAndStore(ctx, fetcher, store, feeds)

	for {
		select {
		case <-ticker.C:
			fetchAndStore(ctx, fetcher, store, feeds)
		case <-ctx.Done():
			fmt.Println("Worker stopped gracefully")
			return
		}
	}
}
