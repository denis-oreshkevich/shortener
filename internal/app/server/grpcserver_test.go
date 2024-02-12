package server

import (
	"context"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	pb "github.com/denis-oreshkevich/shortener/internal/app/server/proto"
	"github.com/denis-oreshkevich/shortener/internal/app/shortener"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestGRPCServer_CreateShortURL(t *testing.T) {
	st := new(storage.MockedStorage)
	sh := shortener.New(st)
	delChannel := make(chan model.BatchDeleteEntry, 3)
	server := NewGRPCServer(sh, config.Get(), delChannel)

	testCases := []struct {
		name   string
		input  *pb.CreateShortURLRequest
		mockOn func(m *storage.MockedStorage) *mock.Call
		assert func(resp *pb.CreateShortURLResponse, err error)
	}{
		{
			name: "Successful #1",
			input: &pb.CreateShortURLRequest{
				UserId: "Denis",
				Url:    "http://testik.test",
			},
			mockOn: func(m *storage.MockedStorage) *mock.Call {
				return m.On("SaveURL", mock.Anything, mock.Anything,
					mock.Anything).Return("AAAAAAAA", nil)
			},
			assert: func(resp *pb.CreateShortURLResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "/AAAAAAAA", resp.Result)
			},
		},
		{
			name: "Error #2",
			input: &pb.CreateShortURLRequest{
				UserId: "Denis",
				Url:    "http://testik.test",
			},
			mockOn: func(m *storage.MockedStorage) *mock.Call {
				return m.On("SaveURL", mock.Anything, mock.Anything,
					mock.Anything).Return("", storage.ErrDBConflict)
			},
			assert: func(resp *pb.CreateShortURLResponse, err error) {
				assert.Error(t, err)
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			mockCall := tt.mockOn(st)
			resp, err := server.CreateShortURL(context.Background(), tt.input)
			st.AssertExpectations(t)
			mockCall.Unset()

			tt.assert(resp, err)
		})
	}
}

func TestGRPCServer_GetByShort(t *testing.T) {
	st := new(storage.MockedStorage)
	sh := shortener.New(st)
	delChannel := make(chan model.BatchDeleteEntry, 3)
	server := NewGRPCServer(sh, config.Get(), delChannel)

	origURL := &storage.OrigURL{OriginalURL: "BBBBBBBB"}

	testCases := []struct {
		name   string
		input  *pb.GetOriginalURLRequest
		mockOn func(m *storage.MockedStorage) *mock.Call
		assert func(resp *pb.GetOriginalURLResponse, err error)
	}{
		{
			name: "Successful #1",
			input: &pb.GetOriginalURLRequest{
				UserId: "Denis",
				Url:    "AAAAAAAA",
			},
			mockOn: func(m *storage.MockedStorage) *mock.Call {
				return m.On("FindURL", mock.Anything, mock.Anything).
					Return(origURL, nil)
			},
			assert: func(resp *pb.GetOriginalURLResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, origURL, resp.OriginalUrl)
			},
		},
		{
			name: "Error #2",
			input: &pb.GetOriginalURLRequest{
				UserId: "Denis",
				Url:    "AAAAAAAA",
			},
			mockOn: func(m *storage.MockedStorage) *mock.Call {
				return m.On("FindURL", mock.Anything, mock.Anything).
					Return(origURL, storage.ErrResultIsDeleted)
			},
			assert: func(resp *pb.GetOriginalURLResponse, err error) {
				assert.Error(t, err)
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			mockCall := tt.mockOn(st)
			resp, err := server.GetByShort(context.Background(), tt.input)
			st.AssertExpectations(t)
			mockCall.Unset()

			tt.assert(resp, err)
		})
	}
}
