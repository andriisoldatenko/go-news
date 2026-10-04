package domain

import "context"

// ArticleReader provides read-only access to articles

type ArticleReader interface {
	GetRecent(n int) []*Article
	GetByID(id string) (*Article, error)
	Close() error
}

// Storage persists and retrieves articles
type Storage interface {
	ArticleReader
	AddArticles(ctx context.Context, articles []*Article) error
}

type SearchableStorage interface {
	Storage
	Searchable
}
