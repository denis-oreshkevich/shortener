package storage

import (
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
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
	fmt.Println("Saved to storage with id =", id, "and value =", url)
	return id
}

func (r *mapStorage) FindURL(id string) (string, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	val, ok := r.items[id]
	fmt.Println("Found in storage by id =", id, "and status =", ok)
	return val, ok
}
