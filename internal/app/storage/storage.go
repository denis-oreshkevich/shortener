package storage

import (
	"context"
	"errors"

	"github.com/denis-oreshkevich/shortener/internal/app/model"
)

var ErrPingNotDB = errors.New("ping not a db storage")
var ErrResultIsDeleted = errors.New("result is deleted")

type Storage interface {
	SaveURL(ctx context.Context, userID string, url string) (string, error)
	SaveURLBatch(ctx context.Context, userID string,
		batch []model.BatchReqEntry) ([]model.BatchRespEntry, error)
	FindURL(ctx context.Context, id string) (*OrigURL, error)

	FindUserURLs(ctx context.Context, userID string) ([]model.URLPair, error)

	DeleteUserURLs(ctx context.Context, bde model.BatchDeleteEntry) error

	Ping(ctx context.Context) error
}
