package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/andriisoldatenko/go-news/domain"
)

type SearchHandlers struct {
	searchable domain.Searchable
}

func NewSearchHandlers(searchable domain.Searchable) *SearchHandlers {
	return &SearchHandlers{searchable: searchable}
}

func (sh *SearchHandlers) HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}
	limitStr := r.URL.Query().Get("limit")
	limit := 10

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	results, err := sh.searchable.Search(r.Context(), query, limit)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"query":   query,
		"results": results,
	})
}
