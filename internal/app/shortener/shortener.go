package shortener

import (
	"context"
	"errors"
	"fmt"

	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"go.uber.org/zap"
)

// ErrUserIsNew indicates that user is new.
var ErrUserIsNew = errors.New("user is new")

// ErrUserItemsNotFound indicates that user URLs not found.
var ErrUserItemsNotFound = errors.New("user items not found")

// TODO think about transactions on this level

// Shortener model represents business logic layer.
type Shortener struct {
	storage storage.Storage
}

// New creates new [*Shortener].
func New(st storage.Storage) *Shortener {
	return &Shortener{
		storage: st,
	}
}

// SaveURL saves URL to storage and returns back short ID.
func (sh *Shortener) SaveURL(ctx context.Context, url string) (string, error) {
	userID, err := sh.GetUserID(ctx)
	if err != nil {
		return "", err
	}
	return sh.storage.SaveURL(ctx, userID, url)
}

// SaveURLBatch saves many URLs to storage and return [[]model.BatchRespEntry] back.
func (sh *Shortener) SaveURLBatch(ctx context.Context,
	batch []model.BatchReqEntry) ([]model.BatchRespEntry, error) {
	userID, err := sh.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	return sh.storage.SaveURLBatch(ctx, userID, batch)
}

// FindURL finds original URL by short ID.
func (sh *Shortener) FindURL(ctx context.Context, id string) (string, error) {
	origURL, err := sh.storage.FindURL(ctx, id)
	if err != nil {
		return "", fmt.Errorf("storage.FindURL. %w", err)
	}
	return origURL.OriginalURL, nil
}

// FindUserURLs finds user's URLs.
func (sh *Shortener) FindUserURLs(ctx context.Context) ([]model.URLPair, error) {
	value := ctx.Value(model.IsUserNew{})
	if value != nil {
		b, ok := value.(bool)
		if !ok {
			return nil, errors.New("IsUserNew is not bool")
		}
		if b {
			return nil, ErrUserIsNew
		}
	}

	userID, err := sh.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	pairs, err := sh.storage.FindUserURLs(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(pairs) == 0 {
		return pairs, ErrUserItemsNotFound
	}
	return pairs, nil
}

// DeleteUserURLs deletes user's URLs.
func (sh *Shortener) DeleteUserURLs(ctx context.Context, in <-chan model.BatchDeleteEntry) {
	for del := range in {
		logger.Log.Debug("received from channel DeleteUserURLs")
		err := sh.storage.DeleteUserURLs(ctx, del)
		if err != nil {
			logger.Log.Error("delete user URLs.", zap.Error(err))
			return
		}
	}
}

// FindStats find statistic by stored values
func (sh *Shortener) FindStats(ctx context.Context) (model.Stat, error) {
	return sh.storage.FindStats(ctx)
}

// Ping pings storage.
func (sh *Shortener) Ping(ctx context.Context) error {
	return sh.storage.Ping(ctx)
}

// GetUserID gets user ID from context.
func (sh *Shortener) GetUserID(ctx context.Context) (string, error) {
	value := ctx.Value(model.UserIDKey{})
	if value == nil {
		return "", errors.New("userID is not present")
	}
	userID, ok := value.(string)
	if !ok {
		return "", errors.New("userID is not string")
	}
	return userID, nil
}
