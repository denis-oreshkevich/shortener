package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapStorage_DeleteUserURLs(t *testing.T) {
	storage := NewMapStorage()
	ctx := context.Background()
	userID := generator.UUIDString()

	shortURL1, err := storage.SaveURL(ctx, userID, "http://localhost:30000/")
	require.NoError(t, err)
	shortURL2, err := storage.SaveURL(ctx, userID, "http://localhost:30001/")
	require.NoError(t, err)

	type args struct {
		ctx context.Context
		bde model.BatchDeleteEntry
	}
	tests := []struct {
		name    string
		storage *MapStorage
		args    args
		assert  func(error)
	}{
		{
			name:    "simple DeleteUserURLs #1",
			storage: storage,
			args: args{
				ctx: ctx,
				bde: model.NewBatchDeleteEntry(userID, []string{shortURL1, shortURL2}),
			},
			assert: func(res error) {
				require.NoError(t, err)
				assert.Len(t, storage.items, 2)
				for k, v := range storage.items {
					assert.True(t, v.DeletedFlag, fmt.Sprintf("element k = %s", k))
				}

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.DeleteUserURLs(tt.args.ctx, tt.args.bde)
			tt.assert(err)
		})
	}
}

func TestMapStorage_FindURL(t *testing.T) {
	storage := NewMapStorage()
	ctx := context.Background()
	userID := generator.UUIDString()

	shortURL, err := storage.SaveURL(ctx, userID, "http://localhost:30000/")
	require.NoError(t, err)

	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		storage *MapStorage
		args    args
		assert  func(*OrigURL, error)
	}{
		{
			name:    "simple FindURL #1",
			storage: storage,
			args: args{
				ctx: ctx,
				id:  shortURL,
			},
			assert: func(res *OrigURL, err error) {
				require.NoError(t, err)
				origURL := NewOrigURL("http://localhost:30000/",
					userID, false)
				assert.Equal(t, &origURL, res)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := storage.FindURL(tt.args.ctx, tt.args.id)
			tt.assert(res, err)
		})
	}
}

func TestMapStorage_FindUserURLs(t *testing.T) {
	storage := NewMapStorage()
	ctx := context.Background()
	userID := generator.UUIDString()

	shortURL1, err := storage.SaveURL(ctx, userID, "http://localhost:30000/")
	require.NoError(t, err)
	shortURL2, err := storage.SaveURL(ctx, userID, "http://localhost:30001/")
	require.NoError(t, err)

	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name   string
		args   args
		assert func([]model.URLPair, error)
	}{
		{
			name: "simple test find user URLs",
			args: args{
				ctx:    ctx,
				userID: userID,
			},
			assert: func(res []model.URLPair, err error) {
				require.NoError(t, err)
				expected := []model.URLPair{
					{
						ShortURL:    fmt.Sprintf("/%s", shortURL1),
						OriginalURL: "http://localhost:30000/",
					},
					{
						ShortURL:    fmt.Sprintf("/%s", shortURL2),
						OriginalURL: "http://localhost:30001/",
					},
				}
				assert.ElementsMatch(t, expected, res)

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := storage.FindUserURLs(tt.args.ctx, tt.args.userID)
			tt.assert(res, err)
		})
	}
}

func TestMapStorage_SaveURL(t *testing.T) {
	storage := NewMapStorage()
	ctx := context.Background()
	userID := generator.UUIDString()

	type args struct {
		ctx    context.Context
		userID string
		url    string
	}
	tests := []struct {
		name   string
		args   args
		assert func(string, error)
	}{
		{
			name: "simple saveURL test #1",
			args: args{
				ctx:    ctx,
				userID: userID,
				url:    "http://localhost:30000/",
			},
			assert: func(shURL string, err error) {
				require.NoError(t, err)
				url, err := storage.FindURL(ctx, shURL)
				require.NoError(t, err)
				assert.Equal(t, "http://localhost:30000/", url.OriginalURL)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := storage.SaveURL(tt.args.ctx, tt.args.userID, tt.args.url)
			tt.assert(res, err)
		})
	}
}

func TestMapStorage_SaveURLBatch(t *testing.T) {
	storage := NewMapStorage()
	ctx := context.Background()
	userID := generator.UUIDString()

	type args struct {
		ctx    context.Context
		userID string
		batch  []model.BatchReqEntry
	}
	tests := []struct {
		name   string
		args   args
		assert func([]model.BatchRespEntry, error)
	}{
		{
			name: "simple SaveURLBatch #1",
			args: args{
				ctx:    ctx,
				userID: userID,
				batch: []model.BatchReqEntry{
					model.NewBatchReqEntry("1", "http://localhost:30000/"),
					model.NewBatchReqEntry("2", "http://localhost:30001/"),
				},
			},
			assert: func(respEntries []model.BatchRespEntry, err error) {
				require.NoError(t, err)
				assert.Len(t, respEntries, 2)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := storage.SaveURLBatch(tt.args.ctx, tt.args.userID, tt.args.batch)
			tt.assert(res, err)
		})
	}
}
