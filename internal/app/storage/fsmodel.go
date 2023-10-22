package storage

type FSModel struct {
	ID          int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
}

func NewFSModel(id int64, shortURL, originalURL, userID string) *FSModel {
	return &FSModel{
		ID:          id,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}
}
