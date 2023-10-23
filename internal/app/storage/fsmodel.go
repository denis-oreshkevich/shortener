package storage

import "github.com/denis-oreshkevich/shortener/internal/app/model"

type FSModel struct {
	ID          int64        `json:"uuid"`
	ShortURL    string       `json:"short_url"`
	OriginalURL string       `json:"original_url"`
	UserID      model.UserID `json:"user_id"`
}

func NewFSModel(id int64, shortURL string, originalURL string, userID model.UserID) *FSModel {
	return &FSModel{
		ID:          id,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}
}
