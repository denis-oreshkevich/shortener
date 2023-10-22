package model

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
)

type URLPair struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewURLPair(id, originalURL string) URLPair {
	baseURL := config.Get().BaseURL()
	return URLPair{
		ShortURL:    fmt.Sprintf("%s/%s", baseURL, id),
		OriginalURL: originalURL,
	}
}
