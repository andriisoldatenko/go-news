package storage

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/andriisoldatenko/go-news/domain"
	"go.etcd.io/bbolt"
)

const articlesBucketName = "articles"

var _ domain.Storage = (*BoltStore)(nil)

type BoltStore struct {
	db *bbolt.DB
}

func NewBoltStore(dbPath string, readOnly bool) (*BoltStore, error) {
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
		Timeout:  1 * time.Second,
		ReadOnly: readOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if !readOnly {
		err = db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(articlesBucketName))
			return err
		})
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &BoltStore{db: db}, nil
}

func (s *BoltStore) Close() error {
	return s.db.Close()
}

func (s *BoltStore) AddArticles(articles []*domain.Article) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(articlesBucketName))
		if bucket == nil {
			return fmt.Errorf("articles bucket not found")
		}

		for _, article := range articles {
			data, err := json.Marshal(article)
			if err != nil {
				return fmt.Errorf("failed to marshal article %s: %w", article.ID, err)
			}
			if err := bucket.Put([]byte(article.ID), data); err != nil {
				return fmt.Errorf("failed to store article %s: %w", article.ID, err)
			}
		}

		return nil

	})
}

func (s *BoltStore) GetRecent(n int) []*domain.Article {
	var articles []*domain.Article
	s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(articlesBucketName))
		if bucket == nil {
			return nil
		}
		bucket.ForEach(func(k, v []byte) error {
			var article domain.Article
			if err := json.Unmarshal(v, &article); err != nil {
				return fmt.Errorf("failed to unmarshal article: %w", err)
			}
			articles = append(articles, &article)
			return nil
		})

		return nil

	})
	// Sort by publication date, newest first
	slices.SortFunc(articles, func(a, b *domain.Article) int {
		if a.Published == nil && b.Published == nil {
			return 0
		}
		if a.Published == nil {
			return 1 // nil published dates sort to the end
		}
		if b.Published == nil {
			return -1 // non-nil comes before nil
		}
		// Compare b to a (reverse order) for newest first
		return b.Published.Compare(*a.Published)
	})
	if n > len(articles) {
		n = len(articles)
	}

	return articles[:n]
}

func (s *BoltStore) GetByID(id string) (*domain.Article, error) {
	var article *domain.Article
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(articlesBucketName))
		if bucket == nil {
			return fmt.Errorf("articles bucket not found")
		}
		data := bucket.Get([]byte(id))
		if data == nil {
			return nil // Article not found
		}
		article = &domain.Article{}
		if err := json.Unmarshal(data, article); err != nil {
			return fmt.Errorf("failed to unmarshal article: %w", err)
		}
		return nil
	})
	return article, err
}
