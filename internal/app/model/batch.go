package model

import (
	"fmt"

	"github.com/denis-oreshkevich/shortener/internal/app/config"
)

// BatchReqEntry model that represents single entry of batch request.
type BatchReqEntry struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// NewBatchReqEntry creates new [BatchReqEntry].
func NewBatchReqEntry(corID string, originalURL string) BatchReqEntry {
	return BatchReqEntry{
		CorrelationID: corID,
		OriginalURL:   originalURL,
	}
}

// BatchRespEntry model that represents single entry of batch response.
type BatchRespEntry struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// NewBatchRespEntry creates new [BatchRespEntry].
func NewBatchRespEntry(corID string, id string) BatchRespEntry {
	baseURL := config.Get().BaseURL()
	return BatchRespEntry{
		CorrelationID: corID,
		ShortURL:      fmt.Sprintf("%s/%s", baseURL, id),
	}
}

// BatchDeleteEntry model that represents single entry of DELETE batch request.
type BatchDeleteEntry struct {
	UserID   string
	ShortIDs []string
}

// NewBatchDeleteEntry creates new [BatchDeleteEntry].
func NewBatchDeleteEntry(userID string, shortIds []string) BatchDeleteEntry {
	return BatchDeleteEntry{
		UserID:   userID,
		ShortIDs: shortIds,
	}
}
