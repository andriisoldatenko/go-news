package reader

import (
	"context"

	"github.com/andriisoldatenko/go-news/domain"
	"github.com/holmes89/gofeed"
)

var _ domain.Fetcher = (*RSSReader)(nil)

type RSSReader struct {
	parser *gofeed.Parser
}

func NewRSSReader() *RSSReader {
	return &RSSReader{
		parser: gofeed.NewParser(),
	}
}

func (r *RSSReader) FetchFeed(ctx context.Context, url string) (*domain.Feed,
	error) {
	feed, err := r.parser.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, err
	}
	domainFeed := &domain.Feed{
		Title:       feed.Title,
		Description: feed.Description,
		Link:        feed.Link,
		Articles:    make([]*domain.Article, 0, len(feed.Items)),
	}

	for _, item := range feed.Items {
		article := &domain.Article{
			ID:          item.GUID,
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			Published:   item.PublishedParsed,
			FeedTitle:   feed.Title,
		}
		domainFeed.Articles = append(domainFeed.Articles, article)
	}

	return domainFeed, nil
}
