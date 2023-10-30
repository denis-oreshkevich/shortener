package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"sync"
)

type MapStorage struct {
	mx sync.RWMutex
	//map userId = slice of URL IDs
	userURLs map[string][]string
	items    map[string]OrigURL
}

var _ Storage = (*MapStorage)(nil)

func NewMapStorage() *MapStorage {
	return &MapStorage{
		userURLs: make(map[string][]string),
		items:    make(map[string]OrigURL),
	}
}

func (ms *MapStorage) SaveURL(ctx context.Context, userID string, url string) (string, error) {
	id := generator.RandString(8)
	ms.mx.Lock()
	defer ms.mx.Unlock()
	ms.saveURLNotSync(id, NewOrigURL(url, userID, false))
	return id, nil
}

func (ms *MapStorage) SaveURLBatch(ctx context.Context, userID string,
	batch []model.BatchReqEntry) ([]model.BatchRespEntry, error) {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	var bResp []model.BatchRespEntry
	for _, b := range batch {
		sh := generator.RandString(8)
		ms.saveURLNotSync(sh, NewOrigURL(b.OriginalURL, userID, false))
		bResp = append(bResp, model.NewBatchRespEntry(b.CorrelationID, sh))
	}
	return bResp, nil
}

func (ms *MapStorage) FindURL(ctx context.Context, id string) (*OrigURL, error) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	val, ok := ms.items[id]
	logger.Log.Debug(fmt.Sprintf("Search in cache by id = %s, and isExist = %t", id, ok))
	if !ok {
		return nil, fmt.Errorf("FindURL value not found by id = %s", id)
	}
	if val.DeletedFlag {
		return nil, ErrResultIsDeleted
	}
	return &val, nil
}

func (ms *MapStorage) FindUserURLs(ctx context.Context, userID string) ([]model.URLPair, error) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	uItems, ok := ms.userURLs[userID]
	logger.Log.Debug(fmt.Sprintf("uItems for userID = %s, isExist = %t", userID, ok))
	if !ok {
		logger.Log.Warn(fmt.Sprintf("uItems for userID = %s, not found", userID))
		return nil, nil
	}
	length := len(uItems)
	var res = make([]model.URLPair, 0, length)
	for i := 0; i < length; i++ {
		id := uItems[i]
		val := ms.items[id]
		p := model.NewURLPair(id, val.OriginalURL)
		res = append(res, p)
	}
	return res, nil
}

func (ms *MapStorage) DeleteUserURLs(ctx context.Context, items []model.BatchDeleteEntry) error {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	return ms.deleteUserURLsNotSync(ctx, items)
}

func (ms *MapStorage) deleteUserURLsNotSync(ctx context.Context,
	items []model.BatchDeleteEntry) error {
	var errs []error
	for _, de := range items {
		_, ok := ms.userURLs[de.UserID]
		if !ok {
			return fmt.Errorf("user URL slice doesnt exist userID = %s", de.UserID)
		}
		for _, shID := range de.ShortIDs {
			url, ok := ms.items[shID]
			if !ok {
				errs = append(errs, fmt.Errorf("shortID doesnt exist userID = %s", de.UserID))
				logger.Log.Debug(fmt.Sprintf("shortID is not present,"+
					"shortID = %s", de.UserID))
				continue
			}
			if de.UserID != url.UserID {
				errs = append(errs, fmt.Errorf("shortID is not of provided user,"+
					"shortID = %s, userI = %s", de.UserID, de.UserID))
				continue
			}
			url.DeletedFlag = true
		}

	}
	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (ms *MapStorage) Ping(ctx context.Context) error {
	return ErrPingNotDB
}

func (ms *MapStorage) saveURLNotSync(id string, orURL OrigURL) {
	uItems, ok := ms.userURLs[orURL.UserID]
	if !ok {
		logger.Log.Debug(fmt.Sprintf("Creating new items map for userID = %s", orURL.UserID))
		uItems = make([]string, 0, 1)
	}
	uItems = append(uItems, id)
	ms.userURLs[orURL.UserID] = uItems
	ms.items[id] = orURL
	logger.Log.Debug(fmt.Sprintf("Saved to cache with userID = %s, id = %s, "+
		"and value = %s", orURL.UserID, id, orURL.OriginalURL))
}
