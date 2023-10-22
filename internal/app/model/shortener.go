package model

type Shortener struct {
	ID          int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewShortener(id int64, shortURL, originalURL string) *Shortener {
	return &Shortener{
		ID:          id,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
}
