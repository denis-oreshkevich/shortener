package shortener

import (
	"context"
	"errors"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
)

var ErrUserIDNotFound = errors.New("userID is not type of string")
var ErrUserItemsNotFound = errors.New("user items not found")

// TODO think about transactions on this level
type Shortener struct {
	storage storage.Storage
}

func New(st storage.Storage) *Shortener {
	return &Shortener{
		storage: st,
	}
}

func (sh *Shortener) SaveURL(ctx context.Context, url string) (string, error) {
	userID, err := sh.getUserID(ctx)
	if err != nil {
		return "", err
	}
	return sh.storage.SaveURL(ctx, userID, url)
}

func (sh *Shortener) SaveURLBatch(ctx context.Context,
	batch []model.BatchReqEntry) ([]model.BatchRespEntry, error) {
	userID, err := sh.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	return sh.storage.SaveURLBatch(ctx, userID, batch)
}
func (sh *Shortener) FindURL(ctx context.Context, id string) (string, error) {
	return sh.storage.FindURL(ctx, id)
}

func (sh *Shortener) FindUserURLs(ctx context.Context) ([]model.URLPair, error) {
	userID, err := sh.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	pairs, err := sh.storage.FindUserURLs(ctx, userID)
	if err != nil {
		return nil, err
	}
	if pairs == nil || len(pairs) == 0 {
		return pairs, ErrUserItemsNotFound
	}
	return pairs, nil
}

func (sh *Shortener) Ping(ctx context.Context) error {
	return sh.storage.Ping(ctx)
}

func (sh *Shortener) getUserID(ctx context.Context) (string, error) {
	value := ctx.Value(model.UserIDKey{})
	if value == nil {
		return "", ErrUserIDNotFound
	}
	userID, ok := value.(string)
	if !ok {
		return "", errors.New("userID is not string")
	}
	return userID, nil
}
