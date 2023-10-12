package model

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

func NewBatchRespEntry(corID string, shortURL string) BatchRespEntry {
	return BatchRespEntry{
		CorrelationID: corID,
		ShortURL:      shortURL,
	}
}
