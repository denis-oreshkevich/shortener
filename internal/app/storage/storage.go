package storage

import (
	"context"
	"errors"

	"github.com/denis-oreshkevich/shortener/internal/app/model"
)

// ErrPingNotDB error happens when you try to Ping not DB storage.
var ErrPingNotDB = errors.New("ping not a db storage")

// ErrResultIsDeleted error happens when you try to get deleted URL.
var ErrResultIsDeleted = errors.New("result is deleted")

// Storage interface for all methods to make communication with repository.
type Storage interface {
	SaveURL(ctx context.Context, userID string, url string) (string, error)
	SaveURLBatch(ctx context.Context, userID string,
		batch []model.BatchReqEntry) ([]model.BatchRespEntry, error)
	FindURL(ctx context.Context, id string) (*OrigURL, error)

	FindUserURLs(ctx context.Context, userID string) ([]model.URLPair, error)

	DeleteUserURLs(ctx context.Context, bde model.BatchDeleteEntry) error

	FindStats(ctx context.Context) (model.Stat, error)

	Ping(ctx context.Context) error
}
