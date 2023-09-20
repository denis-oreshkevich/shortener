package storage

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"sync"
)

type mapStorage struct {
	mx    sync.RWMutex
	items map[string]string
}

func NewMapStorage() Storage {
	return &mapStorage{items: make(map[string]string)}
}

func (r *mapStorage) SaveURL(url string) string {
	id := generator.RandString(8)
	r.mx.Lock()
	defer r.mx.Unlock()
	r.items[id] = url
	logger.Log.Debug(fmt.Sprintf("Saved to storage with id = %s, and value = %s", id, url))
	return id
}

func (r *mapStorage) FindURL(id string) (string, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	val, ok := r.items[id]
	logger.Log.Debug(fmt.Sprintf("Found in storage by id = %s, and status = %t", id, ok))
	return val, ok
}
