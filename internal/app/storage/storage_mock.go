package storage

import (
	"context"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/stretchr/testify/mock"
)

type MockedStorage struct {
	mock.Mock
}

func (m *MockedStorage) FindStats(ctx context.Context) (model.Stat, error) {
	args := m.Called(ctx)
	return args.Get(0).(model.Stat), args.Error(1)
}

func (m *MockedStorage) FindURL(ctx context.Context, id string) (*OrigURL, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*OrigURL), args.Error(1)
}

func (m *MockedStorage) DeleteUserURLs(ctx context.Context, bde model.BatchDeleteEntry) error {
	args := m.Called(ctx, bde)
	return args.Error(0)
}

func (m *MockedStorage) SaveURL(ctx context.Context, userID string, url string) (string, error) {
	args := m.Called(ctx, userID, url)
	return args.String(0), args.Error(1)
}

func (m *MockedStorage) SaveURLBatch(ctx context.Context, userID string,
	batch []model.BatchReqEntry) ([]model.BatchRespEntry, error) {
	args := m.Called(ctx, userID, batch)
	return args.Get(0).([]model.BatchRespEntry), args.Error(1)
}

func (m *MockedStorage) FindUserURLs(ctx context.Context, userID string) ([]model.URLPair, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.URLPair), args.Error(1)
}

func (m *MockedStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
