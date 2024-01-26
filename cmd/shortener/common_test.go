package main

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/server"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"github.com/denis-oreshkevich/shortener/internal/app/util/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var IDURLRegex = regexp.MustCompile(config.Get().BaseURL + "/[A-Za-z0-9]{8}$")

type mockedStorage struct {
	mock.Mock
}

func (m *mockedStorage) FindURL(ctx context.Context, id string) (*storage.OrigURL, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*storage.OrigURL), args.Error(1)
}

func (m *mockedStorage) DeleteUserURLs(ctx context.Context, bde model.BatchDeleteEntry) error {
	args := m.Called(ctx, bde)
	return args.Error(0)
}

func (m *mockedStorage) SaveURL(ctx context.Context, userID string, url string) (string, error) {
	args := m.Called(ctx, userID, url)
	return args.String(0), args.Error(1)
}

func (m *mockedStorage) SaveURLBatch(ctx context.Context, userID string,
	batch []model.BatchReqEntry) ([]model.BatchRespEntry, error) {
	args := m.Called(ctx, userID, batch)
	return args.Get(0).([]model.BatchRespEntry), args.Error(1)
}

func (m *mockedStorage) FindUserURLs(ctx context.Context, userID string) ([]model.URLPair, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.URLPair), args.Error(1)
}

func (m *mockedStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type testConf struct {
	*httptest.Server
	tStorage *mockedStorage
}

func newTestConf(srv *httptest.Server, tStorage *mockedStorage) *testConf {
	return &testConf{Server: srv, tStorage: tStorage}
}

type want struct {
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
	want    want
}

func RunSubTests(t *testing.T, tests []test, testConf *testConf) {
	storage := testConf.tStorage
	client := createHTTPAuthClient(t, testConf.Server)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.reqFunc()

			var mockCall *mock.Call

			if tt.isMock {
				mockCall = tt.mockOn(storage)
			}

			resp, err := client.Do(request)
			require.NoError(t, err)

			if tt.isMock {
				storage.AssertExpectations(t)
				mockCall.Unset()
			}

			defer resp.Body.Close()
			Assert(t, tt, resp)
		})
	}
}

func createHTTPAuthClient(t *testing.T, srv *httptest.Server) *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		require.NoError(t, err)
	}
	token, err := auth.GenerateToken()
	if err != nil {
		require.NoError(t, err)
	}
	cookie := &http.Cookie{
		Name:   server.UserCookieName,
		Value:  token,
		Path:   "",
		Domain: config.Get().ServerAddress,
	}
	c := make([]*http.Cookie, 1)
	c[0] = cookie
	urlStr := srv.URL
	parse, err := url.Parse(urlStr)
	if err != nil {
		require.NoError(t, err)
	}
	jar.SetCookies(parse, c)
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return client
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
		assert.Equal(t, tt.want.body, string(respBody))
	}
}
