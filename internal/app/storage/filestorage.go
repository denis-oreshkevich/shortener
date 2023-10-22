package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

type FileStorage struct {
	mx    sync.RWMutex
	inc   int64
	cache *MapStorage
	file  *os.File
	rw    *bufio.ReadWriter
}

var _ Storage = (*FileStorage)(nil)

func NewFileStorage(filename string) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("NewFileStorage, OpenFile %w", err)
	}
	rw := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))
	items := make(map[string]string)
	var shr = &Shortener{}
	var line int64 = 0
	for {
		data, err := rw.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("NewFileStorage, ReadBytes line #%d %w", line, err)
		}
		err = json.Unmarshal(data, shr)
		if err != nil {
			return nil, fmt.Errorf("NewFileStorage, Unmarshal line #%d %w", line, err)
		}
		items[shr.ShortURL] = shr.OriginalURL
		logger.Log.Debug(fmt.Sprintf("Initializied from file with id = %d, shortURL = %s, OriginalURL = %s", shr.ID, shr.ShortURL, shr.OriginalURL))
		line++
	}
	logger.Log.Info(fmt.Sprintf("Initializing from file count = %d", line))
	cache := NewMapStorage(items)

	return &FileStorage{
		inc:   line,
		cache: cache,
		file:  file,
		rw:    rw,
	}, nil
}

func (fs *FileStorage) SaveURL(url string) (string, error) {
	id := atomic.AddInt64(&fs.inc, 1)
	shURL := generator.RandString(8)
	shorten := newShortener(id, shURL, url)
	marsh, err := json.Marshal(shorten)
	if err != nil {
		return "", fmt.Errorf("fileStorage SaveURL, marshal json %w", err)
	}
	fs.mx.Lock()
	defer fs.mx.Unlock()

	if _, err = fs.rw.Write(marsh); err != nil {
		return "", fmt.Errorf("fileStorage SaveURL, save to file %w", err)
	}
	if err = fs.rw.WriteByte('\n'); err != nil {
		return "", fmt.Errorf("fileStorage SaveURL, write byte %w", err)
	}
	if err = fs.rw.Flush(); err != nil {
		return "", fmt.Errorf("fileStorage SaveURL, flush file %w", err)
	}

	fs.cache.saveURLNotSync(shURL, url)

	return shURL, nil
}

func (fs *FileStorage) FindURL(shortURL string) (string, error) {
	return fs.cache.FindURL(shortURL)
}

func (fs *FileStorage) Close() error {
	return fs.file.Close()
}
