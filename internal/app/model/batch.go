package model

type BatchReqEntry struct {
	CorrelationId string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

func NewBatchReqEntry(corID string, originalURL string) BatchReqEntry {
	return BatchReqEntry{
		CorrelationId: corID,
		OriginalURL:   originalURL,
	}
}

type BatchRespEntry struct {
	CorrelationId string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func NewBatchRespEntry(corID string, shortURL string) BatchRespEntry {
	return BatchRespEntry{
		CorrelationId: corID,
		ShortURL:      shortURL,
	}
}
