package domain

import "context"

// Fetcher retrieves and parses RSS feeds
type Fetcher interface {
	FetchFeed(ctx context.Context, url string) (*Feed, error)
}
