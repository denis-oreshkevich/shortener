package storage

type Storage interface {
	SaveURL(url string) (string, error)
	FindURL(id string) (string, error)
}

type Shortener struct {
	ID          int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func newShortener(id int64, shortURL, originalURL string) *Shortener {
	return &Shortener{
		ID:          id,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
}
