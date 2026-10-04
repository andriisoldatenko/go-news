package domain

import "context"

// SearchResult represents a semantic search match
type SearchResult struct {
	Article *Article
	Score   float32 // Similarity score (0-1, higher = more similar)
}

// Searchable enables semantic search on articles
type Searchable interface {
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
}
