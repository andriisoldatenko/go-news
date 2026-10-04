package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/andriisoldatenko/go-news/domain"
)

var _ domain.SearchableStorage = (*SearchStore)(nil)

type SearchStore struct {
	domain.Storage
	embedder domain.Embedder
	qdrant   *QdrantClient
}

func NewSearchStore(storage domain.Storage, embedder domain.Embedder, qdrant *QdrantClient) *SearchStore {
	return &SearchStore{
		Storage:  storage,
		embedder: embedder,
		qdrant:   qdrant,
	}
}

func NewSearchStoreWithDefaults(ctx context.Context, storage domain.Storage) (*SearchStore, error) {
	embedder := NewOllamaEmbedder("", "")
	qdrant, err := NewQdrantClient(ctx, "", "")
	if err != nil {
		return nil, fmt.Errorf("connect to qdrant: %w", err)
	}
	return NewSearchStore(storage, embedder, qdrant), nil
}

func (ss *SearchStore) AddArticles(ctx context.Context, articles []*domain.Article) error {
	if err := ss.Storage.AddArticles(ctx, articles); err != nil {
		return err
	}
	for _, article := range articles {
		if err := ss.embedAndStore(ctx, article); err != nil {
			log.Printf("Warning: Failed to embed article %s: %v", article.ID,
				err)
			// Continue - embedding failures don't break article storage
		}
	}
	return nil
}

func (ss *SearchStore) embedAndStore(ctx context.Context, article *domain.Article) error {
	text := article.Title
	if article.Description != "" {
		text += "\n\n" + article.Description
	}
	vector, err := ss.embedder.Embed(ctx, text)
	if err != nil {
		return fmt.Errorf("embed text: %w", err)
	}

	metadata := map[string]any{
		"title":      article.Title,
		"feed_title": article.FeedTitle}
	return ss.qdrant.Store(ctx, article.ID, vector, metadata)

}
func (ss *SearchStore) Search(ctx context.Context, query string, limit int) ([]domain.SearchResult, error) {
	queryVector, err := ss.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}
	matches, err := ss.qdrant.Search(ctx, queryVector, limit)
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}
	results := make([]domain.SearchResult, 0, len(matches))
	for _, match := range matches {
		article, err := ss.Storage.GetByID(match.ID)
		if err != nil {
			log.Printf("Warning: Could not fetch article %s: %v", match.ID, err)
			continue
		}
		results = append(results, domain.SearchResult{
			Article: article,
			Score:   match.Score,
		})
	}
	return results, nil
}
