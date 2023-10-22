package storage

import (
	"context"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
)

type Storage interface {
	SaveURL(ctx context.Context, url string) (string, error)
	SaveURLBatch(ctx context.Context, batch []model.BatchReqEntry) ([]model.BatchRespEntry, error)
	FindURL(ctx context.Context, id string) (string, error)
}
