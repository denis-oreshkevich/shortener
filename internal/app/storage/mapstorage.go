package storage

import (
	"fmt"
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

func (r *MapStorage) SaveURL(url string) (string, error) {
	id := generator.RandString(8)
	r.saveURL(id, url)
	return id, nil
}

func (r *MapStorage) FindURL(id string) (string, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	val, ok := r.items[id]
	logger.Log.Debug(fmt.Sprintf("Found in cache by id = %s, and isExist = %t", id, ok))
	if !ok {
		return val, fmt.Errorf("FindURL value not found by id = %s", id)
	}
	return val, nil
}

func (r *MapStorage) saveURL(id, url string) {
	r.mx.Lock()
	defer r.mx.Unlock()
	r.items[id] = url
	logger.Log.Debug(fmt.Sprintf("Saved to cache with id = %s, and value = %s", id, url))
}
