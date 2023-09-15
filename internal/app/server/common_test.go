package server

import (
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

var IDURLRegex = regexp.MustCompile(config.Get().BaseURL() + "/[A-Za-z]{8}")

type mockedStorage struct {
	mock.Mock
}

func (m *mockedStorage) SaveURL(url string) string {
	args := m.Called(url)
	return args.String(0)
}

func (m *mockedStorage) FindURL(id string) (string, bool) {
	args := m.Called(id)
	return args.String(0), args.Bool(1)
}

type Want struct {
	contentType    string
	statusCode     int
	body           string
	headerLocation string
}

type test struct {
	name    string
	isMock  bool
	mockOn  func(m *mockedStorage) *mock.Call
	reqFunc func() *http.Request
	want    Want
}

func RunSubTests(t *testing.T, tests []test) {
	tStorage := new(mockedStorage)
	conf := config.Get()
	uh := handler.New(conf, tStorage)
	srv := New(conf, uh)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.reqFunc()
			w := httptest.NewRecorder()

			var mockCall *mock.Call

			if tt.isMock {
				mockCall = tt.mockOn(tStorage)
			}

			srv.router.ServeHTTP(w, request)

			if tt.isMock {
				tStorage.AssertExpectations(t)
				mockCall.Unset()
			}
			result := w.Result()
			defer result.Body.Close()
			Assert(t, tt, result)
		})
	}
}

func Assert(t *testing.T, tt test, result *http.Response) {
	assert.Equal(t, tt.want.statusCode, result.StatusCode)
	assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
	if tt.want.headerLocation != "" {
		assert.Equal(t, tt.want.headerLocation, result.Header.Get("Location"))
	}
	if tt.want.body != "" {
		respBody, err := io.ReadAll(result.Body)
		require.NoError(t, err)
		assert.True(t, IDURLRegex.MatchString(string(respBody)))
	}
}
