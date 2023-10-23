package storage

import (
	"context"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"sync"
)

type MapStorage struct {
	mx sync.RWMutex
	//map userId = slice of URL IDs
	userURLs map[model.UserID][]string
	items    map[string]string
}

var _ Storage = (*MapStorage)(nil)

func NewMapStorage() *MapStorage {
	return &MapStorage{
		userURLs: make(map[model.UserID][]string),
		items:    make(map[string]string),
	}
}

func (ms *MapStorage) SaveURL(ctx context.Context, userID model.UserID, url string) (string, error) {
	id := generator.RandString(8)
	ms.mx.Lock()
	defer ms.mx.Unlock()
	ms.saveURLNotSync(userID, id, url)
	return id, nil
}

func (ms *MapStorage) SaveURLBatch(ctx context.Context, userID model.UserID,
	batch []model.BatchReqEntry) ([]model.BatchRespEntry, error) {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	var bResp []model.BatchRespEntry
	for _, b := range batch {
		sh := generator.RandString(8)
		ms.saveURLNotSync(userID, sh, b.OriginalURL)
		bResp = append(bResp, model.NewBatchRespEntry(b.CorrelationID, sh))
	}
	return bResp, nil
}

func (ms *MapStorage) FindURL(ctx context.Context, id string) (string, error) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	val, ok := ms.items[id]
	logger.Log.Debug(fmt.Sprintf("Search in cache by id = %s, and isExist = %t", id, ok))
	if !ok {
		return val, fmt.Errorf("FindURL value not found by id = %s", id)
	}
	return val, nil
}

func (ms *MapStorage) FindUserURLs(ctx context.Context, userID model.UserID) ([]model.URLPair, error) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	uItems, ok := ms.userURLs[userID]
	logger.Log.Debug(fmt.Sprintf("uItems for userID = %s, isExist = %t", userID, ok))
	if !ok {
		return nil, fmt.Errorf("uItems not found by userID = %s", userID)
	}
	length := len(uItems)
	var res = make([]model.URLPair, length)
	for i := 0; i < length; i++ {
		id := uItems[i]
		val := ms.items[id]
		p := model.NewURLPair(id, val)
		res = append(res, p)
	}

	return res, nil
}

func (ms *MapStorage) saveURLNotSync(userID model.UserID, id, url string) {
	uItems, ok := ms.userURLs[userID]
	if !ok {
		logger.Log.Debug(fmt.Sprintf("Creating new items map for userID = %s", userID))
		uItems = make([]string, 0, 1)
	}
	uItems = append(uItems, id)
	ms.userURLs[userID] = uItems
	ms.items[id] = url
	logger.Log.Debug(fmt.Sprintf("Saved to cache with userID = %s, id = %s, "+
		"and value = %s", userID, id, url))
}

func (ms *MapStorage) Ping(ctx context.Context) error {
	return ErrPingNotDB
}
