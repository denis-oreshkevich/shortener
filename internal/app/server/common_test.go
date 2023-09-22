package server

import (
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

var IDURLRegex = regexp.MustCompile(config.Get().BaseURL() + "/[A-Za-z]{8}$")

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

type testSrv struct {
	*httptest.Server
	tStorage *mockedStorage
}

func newTestSrv(srv *httptest.Server, tStorage *mockedStorage) *testSrv {
	return &testSrv{Server: srv, tStorage: tStorage}
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

func RunSubTests(t *testing.T, tests []test, srv *testSrv) {
	storage := srv.tStorage
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
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
