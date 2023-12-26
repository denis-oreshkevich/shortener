package storage

// FSModel model that stores in file.
type FSModel struct {
	ID          int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	DeletedFlag bool   `db:"is_deleted"`
}

// NewFSModel creates new [FSModel].
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

// OrigURL model.
type OrigURL struct {
	OriginalURL string
	UserID      string
	DeletedFlag bool
}

// NewOrigURL creates new [OrigURL].
func NewOrigURL(originalURL string, userID string, delFlag bool) OrigURL {
	return OrigURL{
		OriginalURL: originalURL,
		UserID:      userID,
		DeletedFlag: delFlag,
	}
}
