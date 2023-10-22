package storage

type FSModel struct {
	ID          int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFSModel(id int64, shortURL, originalURL string) *FSModel {
	return &FSModel{
		ID:          id,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
}
