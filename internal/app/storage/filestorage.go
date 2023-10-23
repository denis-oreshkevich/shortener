package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
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
	cache := NewMapStorage()
	var shr = &FSModel{}
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
		cache.saveURLNotSync(model.UserID(shr.UserID), shr.ShortURL, shr.OriginalURL)
		logger.Log.Debug(fmt.Sprintf("Initializied from file with id = %d, shortURL = %s, OriginalURL = %s", shr.ID, shr.ShortURL, shr.OriginalURL))
		line++
	}
	logger.Log.Info(fmt.Sprintf("Initializing from file count = %d", line))
	return &FileStorage{
		inc:   line,
		cache: cache,
		file:  file,
		rw:    rw,
	}, nil
}

func (fs *FileStorage) SaveURL(ctx context.Context, userID model.UserID, url string) (string, error) {
	id := atomic.AddInt64(&fs.inc, 1)
	shURL := generator.RandString(8)
	shorten := NewFSModel(id, shURL, url, userID)
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

	fs.cache.saveURLNotSync(userID, shURL, url)

	return shURL, nil
}

func (fs *FileStorage) SaveURLBatch(ctx context.Context, userID model.UserID,
	batch []model.BatchReqEntry) ([]model.BatchRespEntry, error) {
	var bResp []model.BatchRespEntry
	fs.mx.Lock()
	defer fs.mx.Unlock()
	for _, b := range batch {
		id := atomic.AddInt64(&fs.inc, 1)
		shURL := generator.RandString(8)
		shorten := NewFSModel(id, shURL, b.OriginalURL, userID)
		marsh, err := json.Marshal(shorten)
		if err != nil {
			return nil, fmt.Errorf("fileStorage SaveURLBatch, marshal json %w", err)
		}

		if _, err = fs.rw.Write(marsh); err != nil {
			return nil, fmt.Errorf("fileStorage SaveURLBatch. save to file %w", err)
		}
		if err = fs.rw.WriteByte('\n'); err != nil {
			return nil, fmt.Errorf("fileStorage SaveURLBatch. write byte %w", err)
		}
		fs.cache.saveURLNotSync(userID, shURL, b.OriginalURL)
		resp := model.NewBatchRespEntry(b.CorrelationID, shURL)
		bResp = append(bResp, resp)
	}
	if err := fs.rw.Flush(); err != nil {
		return nil, fmt.Errorf("fileStorage SaveURLBatch. flush file %w", err)
	}
	return bResp, nil
}

func (fs *FileStorage) FindURL(ctx context.Context, id string) (string, error) {
	return fs.cache.FindURL(ctx, id)
}

func (fs *FileStorage) FindUserURLs(ctx context.Context, userID model.UserID) ([]model.URLPair, error) {
	return fs.cache.FindUserURLs(ctx, userID)
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	return ErrPingNotDB
}

func (fs *FileStorage) Close() error {
	return fs.file.Close()
}
