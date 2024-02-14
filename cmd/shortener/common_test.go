package main

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/server"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"github.com/denis-oreshkevich/shortener/internal/app/util/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var IDURLRegex = regexp.MustCompile(config.Get().BaseURL() + "/[A-Za-z0-9]{8}$")

type testConf struct {
	*httptest.Server
	tStorage *storage.MockedStorage
}

func newTestConf(srv *httptest.Server, tStorage *storage.MockedStorage) *testConf {
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
	mockOn  func(m *storage.MockedStorage) *mock.Call
	reqFunc func() *http.Request
	want    want
}

func RunSubTests(t *testing.T, tests []test, testConf *testConf) {
	st := testConf.tStorage
	client := createHTTPAuthClient(t, testConf.Server)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.reqFunc()

			var mockCall *mock.Call

			if tt.isMock {
				mockCall = tt.mockOn(st)
			}

			resp, err := client.Do(request)
			require.NoError(t, err)

			if tt.isMock {
				st.AssertExpectations(t)
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
		Domain: config.Get().Host(),
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
