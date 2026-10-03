package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/andriisoldatenko/go-news/domain"
)

type Handlers struct {
	articles domain.ArticleReader
}

func New(articles domain.ArticleReader) *Handlers {
	return &Handlers{
		articles: articles,
	}
}

func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/articles", h.articlesHandler)
}
func (h *Handlers) articlesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return

	}
	n := 10
	if countStr := r.URL.Query().Get("count"); countStr != "" {
		if count, err := strconv.Atoi(countStr); err == nil && count > 0 {
			n = count
		}
	}

	articles := h.articles.GetRecent(n)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(articles); err != nil {
		http.Error(w, "failed to encode response", http.
			StatusInternalServerError)
		return
	}
}
