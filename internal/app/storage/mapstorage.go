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
	mx    sync.RWMutex
	items map[string]string
}

var _ Storage = (*MapStorage)(nil)

func NewMapStorage(items map[string]string) *MapStorage {
	return &MapStorage{items: items}
}

func (ms *MapStorage) SaveURL(ctx context.Context, url string) (string, error) {
	id := generator.RandString(8)
	ms.mx.Lock()
	defer ms.mx.Unlock()
	ms.saveURLNotSync(id, url)
	return id, nil
}

func (ms *MapStorage) SaveURLBatch(ctx context.Context, batch []model.BatchReqEntry) ([]model.BatchRespEntry, error) {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	var bResp []model.BatchRespEntry
	for _, b := range batch {
		sh := generator.RandString(8)
		ms.saveURLNotSync(sh, b.OriginalURL)
		bResp = append(bResp, model.NewBatchRespEntry(b.CorrelationID, sh))
	}
	return bResp, nil
}

func (ms *MapStorage) FindURL(ctx context.Context, id string) (string, error) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	val, ok := ms.items[id]
	logger.Log.Debug(fmt.Sprintf("Found in cache by id = %s, and isExist = %t", id, ok))
	if !ok {
		return val, fmt.Errorf("FindURL value not found by id = %s", id)
	}
	return val, nil
}

func (ms *MapStorage) saveURLNotSync(id, url string) {
	ms.items[id] = url
	logger.Log.Debug(fmt.Sprintf("Saved to cache with id = %s, and value = %s", id, url))
}
