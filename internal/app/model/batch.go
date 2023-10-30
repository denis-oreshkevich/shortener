package model

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
)

type BatchReqEntry struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

func NewBatchReqEntry(corID string, originalURL string) BatchReqEntry {
	return BatchReqEntry{
		CorrelationID: corID,
		OriginalURL:   originalURL,
	}
}

type BatchRespEntry struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func NewBatchRespEntry(corID string, id string) BatchRespEntry {
	baseURL := config.Get().BaseURL()
	return BatchRespEntry{
		CorrelationID: corID,
		ShortURL:      fmt.Sprintf("%s/%s", baseURL, id),
	}
}

type BatchDeleteEntry struct {
	UserID   string
	ShortIDs []string
}

func NewBatchDeleteEntry(userID string, shortIds []string) BatchDeleteEntry {
	return BatchDeleteEntry{
		UserID:   userID,
		ShortIDs: shortIds,
	}
}
