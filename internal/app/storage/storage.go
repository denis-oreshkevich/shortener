package storage

import (
	"context"
	"errors"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
)

var ErrPingNotDB = errors.New("ping not a db storage")

type Storage interface {
	SaveURL(ctx context.Context, userID string, url string) (string, error)
	SaveURLBatch(ctx context.Context, userID string,
		batch []model.BatchReqEntry) ([]model.BatchRespEntry, error)
	FindURL(ctx context.Context, id string) (string, error)

	FindUserURLs(ctx context.Context, userID string) ([]model.URLPair, error)

	Ping(ctx context.Context) error
}
