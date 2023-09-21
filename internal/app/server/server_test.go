package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/handler"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPost(t *testing.T) {
	conf := config.Get()
	postTests := []test{
		{
			name:   "simple Post test #1",
			isMock: true,
			mockOn: func(m *mockedStorage) *mock.Call {
				return m.On("SaveURL", mock.Anything).Return("CCCCCCCC")
			},
			reqFunc: func() *http.Request {
				body := strings.NewReader("https://practicum.yandex.ru/")
				return httptest.NewRequest("POST", "/", body)
			},
			want: Want{
				contentType: handler.TextPlain,
				statusCode:  201,
				body:        conf.BaseURL() + "/" + "CCCCCCCC",
			},
		},
		{
			name:   "nil body Post test #2",
			isMock: false,
			reqFunc: func() *http.Request {
				return httptest.NewRequest("POST", "/", nil)
			},
			want: Want{
				contentType: handler.TextPlain,
				statusCode:  400,
			},
		},
		{
			name:   "bad body url Post test #3",
			isMock: false,
			reqFunc: func() *http.Request {
				body := strings.NewReader("ahahahah")
				return httptest.NewRequest("POST", "/", body)
			},
			want: Want{
				contentType: handler.TextPlain,
				statusCode:  400,
			},
		},
	}
	RunSubTests(t, postTests)
}

func TestGet(t *testing.T) {
	conf := config.Get()
	getTests := []test{
		{
			name:   "simple Get test #1",
			isMock: true,
			mockOn: func(m *mockedStorage) *mock.Call {
				return m.On("FindURL", "AAAAAAAA").Return("http://localhost:30001/", true)
			},
			reqFunc: func() *http.Request {
				return httptest.NewRequest("GET", conf.BasePath()+"/AAAAAAAA", nil)
			},
			want: Want{
				contentType:    handler.TextPlain,
				statusCode:     307,
				headerLocation: "http://localhost:30001/",
			},
		},
		{
			name:   "bad id Get test #2",
			isMock: false,
			reqFunc: func() *http.Request {
				return httptest.NewRequest("GET", conf.BasePath()+"/HHH", nil)
			},
			want: Want{
				contentType: handler.TextPlain,
				statusCode:  400,
			},
		},
		{
			name:   "not stored url Get test #3",
			isMock: true,
			mockOn: func(m *mockedStorage) *mock.Call {
				return m.On("FindURL", "HHHHHHHH").Return("", false)
			},
			reqFunc: func() *http.Request {
				return httptest.NewRequest("GET", conf.BasePath()+"/HHHHHHHH", nil)
			},
			want: Want{
				contentType: handler.TextPlain,
				statusCode:  400,
			},
		},
	}
	RunSubTests(t, getTests)
}

func TestShortenPost(t *testing.T) {
	conf := config.Get()
	res, err := json.Marshal(handler.NewResult(conf.BaseURL() + "/" + "EEEEEEEE"))
	if err != nil {
		logger.Log.Error("marshal json", zap.Error(err))
	}
	tests := []test{
		{
			name:   "simple ShortenPost test #1",
			isMock: true,
			mockOn: func(m *mockedStorage) *mock.Call {
				return m.On("SaveURL", mock.Anything).Return("EEEEEEEE")
			},
			reqFunc: func() *http.Request {
				url := handler.NewURL("https://practicum.yandex.ru/")
				marshal, err := json.Marshal(url)
				if err != nil {
					logger.Log.Error("marshal json", zap.Error(err))
				}
				return httptest.NewRequest("POST", "/api/shorten", bytes.NewReader(marshal))
			},
			want: Want{
				contentType: handler.ApplicationJSON,
				statusCode:  201,
				body:        string(res),
			},
		},
		{
			name:   "empty body url ShortenPost test #2",
			isMock: false,
			reqFunc: func() *http.Request {
				url := handler.NewURL("")
				marshal, err := json.Marshal(url)
				if err != nil {
					logger.Log.Error("marshal json", zap.Error(err))
				}
				return httptest.NewRequest("POST", "/api/shorten", bytes.NewReader(marshal))
			},
			want: Want{
				contentType: handler.TextPlain,
				statusCode:  400,
			},
		},
		{
			name:   "empty json ShortenPost test #3",
			isMock: false,
			reqFunc: func() *http.Request {
				body := strings.NewReader("{}")
				return httptest.NewRequest("POST", "/api/shorten", body)
			},
			want: Want{
				contentType: handler.TextPlain,
				statusCode:  400,
			},
		},
	}
	RunSubTests(t, tests)
}

func TestNoRoutes(t *testing.T) {
	conf := config.Get()
	tests := []test{
		{
			name:   "bad url Get test '/test/AAAAAAAA' #1",
			isMock: false,
			reqFunc: func() *http.Request {
				return httptest.NewRequest("GET", conf.BasePath()+"/test/AAAAAAAA", nil)
			},
			want: Want{
				contentType: handler.TextPlain,
				statusCode:  400,
			},
		},
		{
			name:   "bad url Post test '/test' #2",
			isMock: false,
			reqFunc: func() *http.Request {
				body := strings.NewReader("https://practicum.yandex.ru/")
				return httptest.NewRequest("POST", "/test", body)
			},
			want: Want{
				contentType: handler.TextPlain,
				statusCode:  400,
			},
		},
		{
			name:   "bad method test 'DELETE' #3",
			isMock: false,
			reqFunc: func() *http.Request {
				body := strings.NewReader("https://practicum.yandex.ru/")
				return httptest.NewRequest("DELETE", "/", body)
			},
			want: Want{
				contentType: handler.TextPlain,
				statusCode:  400,
			},
		},
	}
	RunSubTests(t, tests)
}

func TestGzipCompression(t *testing.T) {
	conf := config.Get()
	tStorage := new(mockedStorage)
	tStorage.On("SaveURL", mock.Anything).Return("MMMMMMMM")
	uh := handler.New(conf, tStorage)
	s := New(conf, uh)

	srv := httptest.NewServer(s.router)
	defer srv.Close()

	requestBody := `{
    	"url": "https://practicum.yandex.ru/"
	}`

	successBody := `{
        "result":"` + conf.BaseURL() + "/MMMMMMMM" + `"
	}`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL+"/api/shorten", buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL+"/api/shorten", buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
}
