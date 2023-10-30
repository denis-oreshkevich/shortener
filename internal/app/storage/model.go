package storage

type FSModel struct {
	ID          int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	DeletedFlag bool   `db:"is_deleted"`
}

func NewFSModel(id int64, shortURL string,
	originalURL string, userID string, delFlag bool) *FSModel {
	return &FSModel{
		ID:          id,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
		DeletedFlag: delFlag,
	}
}

type OrigURL struct {
	OriginalURL string
	UserID      string
	DeletedFlag bool
}

func NewOrigURL(originalURL string, userID string, delFlag bool) OrigURL {
	return OrigURL{
		OriginalURL: originalURL,
		UserID:      userID,
		DeletedFlag: delFlag,
	}
}
