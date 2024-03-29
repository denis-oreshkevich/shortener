package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
)

// FileStorage file storage.
type FileStorage struct {
	filename string
	mx       sync.RWMutex
	inc      int64
	cache    *MapStorage
	file     *os.File
	rw       *bufio.ReadWriter
}

var _ Storage = (*FileStorage)(nil)

// NewFileStorage creates new [*FileStorage].
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
		cache.saveURLNotSync(shr.ShortURL, NewOrigURL(shr.OriginalURL, shr.UserID, shr.DeletedFlag))
		logger.Log.Debug(fmt.Sprintf("Initializied from file with id = %d, shortURL = %s, OriginalURL = %s", shr.ID, shr.ShortURL, shr.OriginalURL))
		line++
	}
	logger.Log.Info(fmt.Sprintf("Initializing from file count = %d", line))
	return &FileStorage{
		filename: filename,
		inc:      line,
		cache:    cache,
		file:     file,
		rw:       rw,
	}, nil
}

// SaveURL saves original URL to file and map and returns short URL.
func (fs *FileStorage) SaveURL(ctx context.Context, userID string, url string) (string, error) {
	id := atomic.AddInt64(&fs.inc, 1)
	shURL := generator.RandString()
	shorten := NewFSModel(id, shURL, url, userID, false)
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

	fs.cache.saveURLNotSync(shURL, NewOrigURL(url, userID, false))

	return shURL, nil
}

// SaveURLBatch saves many URLs to file and map and return [[]model.BatchRespEntry] back.
func (fs *FileStorage) SaveURLBatch(ctx context.Context, userID string,
	batch []model.BatchReqEntry) ([]model.BatchRespEntry, error) {
	var bResp []model.BatchRespEntry
	fs.mx.Lock()
	defer fs.mx.Unlock()
	for _, b := range batch {
		id := atomic.AddInt64(&fs.inc, 1)
		shURL := generator.RandString()
		shorten := NewFSModel(id, shURL, b.OriginalURL, userID, false)
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
		fs.cache.saveURLNotSync(shURL, NewOrigURL(b.OriginalURL, userID, false))
		resp := model.NewBatchRespEntry(b.CorrelationID, shURL)
		bResp = append(bResp, resp)
	}
	if err := fs.rw.Flush(); err != nil {
		return nil, fmt.Errorf("fileStorage SaveURLBatch. flush file %w", err)
	}
	return bResp, nil
}

// FindURL finds original URL in file and map by short ID.
func (fs *FileStorage) FindURL(ctx context.Context, id string) (*OrigURL, error) {
	return fs.cache.FindURL(ctx, id)
}

// FindUserURLs finds user's URLs in file and map.
func (fs *FileStorage) FindUserURLs(ctx context.Context, userID string) ([]model.URLPair, error) {
	return fs.cache.FindUserURLs(ctx, userID)
}

// DeleteUserURLs deletes user's URLs.
func (fs *FileStorage) DeleteUserURLs(ctx context.Context, bde model.BatchDeleteEntry) error {
	fs.mx.Lock()
	defer fs.mx.Unlock()
	content := make(map[string]*FSModel)
	var shr = &FSModel{}
	err := fs.readFileToMap(shr, content)
	if err != nil {
		return fmt.Errorf("read file to map. %w", err)
	}
	var errs []error
	ids := bde.ShortIDs
	for _, id := range ids {
		fsm, ok := content[id]
		if !ok {
			errs = append(errs, fmt.Errorf("shortID = %s, is not exist", id))
			continue
		}
		if fsm.UserID != bde.UserID {
			errs = append(errs, fmt.Errorf("shortID = %s is not of userID = %s ",
				id, bde.UserID))
			continue
		}
		fsm.DeletedFlag = true
	}
	err = fs.file.Truncate(0)
	if err != nil {
		return fmt.Errorf("truncate file. %w", err)
	}
	if _, err = fs.file.Seek(0, 0); err != nil {
		return fmt.Errorf("seek file. %w", err)
	}
	fs.cache = NewMapStorage()
	err = fs.writeContentToCacheAndFile(content)
	if err != nil {
		return fmt.Errorf("write content to cache and file. %w", err)
	}
	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

// FindStats finds statistic by saved requests.
func (fs *FileStorage) FindStats(ctx context.Context) (model.Stat, error) {
	return fs.cache.FindStats(ctx)
}

func (fs *FileStorage) readFileToMap(shr *FSModel, content map[string]*FSModel) error {
	for {
		data, err := fs.rw.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("can't read file. %w", err)
		}
		err = json.Unmarshal(data, shr)
		if err != nil {
			return fmt.Errorf("unmarshal JSON from file %w", err)
		}
		content[shr.ShortURL] = shr
	}
	return nil
}

func (fs *FileStorage) writeContentToCacheAndFile(content map[string]*FSModel) error {
	for _, cont := range content {
		marsh, err := json.Marshal(cont)
		if err != nil {
			return fmt.Errorf("marshal json %w", err)
		}

		if _, err = fs.rw.Write(marsh); err != nil {
			return fmt.Errorf("save to file %w", err)
		}
		if err = fs.rw.WriteByte('\n'); err != nil {
			return fmt.Errorf("write byte %w", err)
		}
		fs.cache.saveURLNotSync(cont.ShortURL, NewOrigURL(cont.OriginalURL,
			cont.UserID, cont.DeletedFlag))
	}
	if err := fs.rw.Flush(); err != nil {
		return fmt.Errorf("flush file %w", err)
	}
	return nil
}

// Ping Returns an error.
func (fs *FileStorage) Ping(ctx context.Context) error {
	return ErrPingNotDB
}

// Close closes file.
func (fs *FileStorage) Close() error {
	return fs.file.Close()
}
