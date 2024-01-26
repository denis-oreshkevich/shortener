package model

import (
	"fmt"

	"github.com/denis-oreshkevich/shortener/internal/app/config"
)

// URLPair model represents short and original URLs.
type URLPair struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// NewURLPair creates new [URLPair].
func NewURLPair(id, originalURL string) URLPair {
	baseURL := config.Get().BaseURL()
	return URLPair{
		ShortURL:    fmt.Sprintf("%s/%s", baseURL, id),
		OriginalURL: originalURL,
	}
}
