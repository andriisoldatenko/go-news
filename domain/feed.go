package domain

// Feed represents an RSS/Atom feed
type Feed struct {
	Title       string
	Description string
	Link        string
	Articles    []*Article
}
