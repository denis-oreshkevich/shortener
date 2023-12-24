package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/server"
	"github.com/denis-oreshkevich/shortener/internal/app/shortener"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPost(t *testing.T) {
	conf := config.Get()
	tStorage := new(mockedStorage)
	short := shortener.New(tStorage)
	delChannel := make(chan model.BatchDeleteEntry, 3)
	uh := server.New(conf, short, delChannel)
	r := SetUpRouter(conf, uh)

	srv := httptest.NewServer(r)
	defer srv.Close()
	tSrv := newTestConf(srv, tStorage)

	postTests := []test{
		{
			name:   "simple Post test #1",
			isMock: true,
			mockOn: func(m *mockedStorage) *mock.Call {
				return m.On("SaveURL", mock.Anything, mock.Anything,
					mock.Anything).Return("CCCCCCCC", nil)
			},
			reqFunc: func() *http.Request {
				body := strings.NewReader("https://practicum.yandex.ru/")
				req := httptest.NewRequest("POST", srv.URL+"/", body)
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType: server.TextPlain,
				statusCode:  201,
				body:        conf.BaseURL() + "/" + "CCCCCCCC",
			},
		},
		{
			name:   "nil body Post test #2",
			isMock: false,
			reqFunc: func() *http.Request {
				req := httptest.NewRequest("POST", srv.URL+"/", nil)
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType: server.TextPlain,
				statusCode:  400,
			},
		},
		{
			name:   "bad body url Post test #3",
			isMock: false,
			reqFunc: func() *http.Request {
				body := strings.NewReader("ahahahah")
				req := httptest.NewRequest("POST", srv.URL+"/", body)
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType: server.TextPlain,
				statusCode:  400,
			},
		},
	}
	RunSubTests(t, postTests, tSrv)
}

func TestGet(t *testing.T) {
	conf := config.Get()
	tStorage := new(mockedStorage)
	short := shortener.New(tStorage)
	delChannel := make(chan model.BatchDeleteEntry, 3)
	uh := server.New(conf, short, delChannel)
	r := SetUpRouter(conf, uh)

	srv := httptest.NewServer(r)
	defer srv.Close()
	tSrv := newTestConf(srv, tStorage)

	getTests := []test{
		{
			name:   "simple Get test #1",
			isMock: true,
			mockOn: func(m *mockedStorage) *mock.Call {
				origURL := storage.NewOrigURL("http://localhost:30001/", "", false)
				return m.On("FindURL", mock.Anything, "AAAAAAAA").Return(
					&origURL, nil)
			},
			reqFunc: func() *http.Request {
				req := httptest.NewRequest("GET", srv.URL+conf.BasePath()+"/AAAAAAAA", nil)
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType:    server.TextPlain,
				statusCode:     307,
				headerLocation: "http://localhost:30001/",
			},
		},
		{
			name:   "bad id Get test #2",
			isMock: false,
			reqFunc: func() *http.Request {
				req := httptest.NewRequest("GET", srv.URL+conf.BasePath()+"/HHH", nil)
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType: server.TextPlain,
				statusCode:  400,
			},
		},
		{
			name:   "not stored url Get test #3",
			isMock: true,
			mockOn: func(m *mockedStorage) *mock.Call {
				return m.On("FindURL", mock.Anything, "HHHHHHHH").Return(&storage.OrigURL{}, errors.New("test error"))
			},
			reqFunc: func() *http.Request {
				req := httptest.NewRequest("GET", srv.URL+conf.BasePath()+"/HHHHHHHH", nil)
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType: server.TextPlain,
				statusCode:  400,
			},
		},
	}
	RunSubTests(t, getTests, tSrv)
}

func TestShortenPost(t *testing.T) {
	conf := config.Get()
	tStorage := new(mockedStorage)
	short := shortener.New(tStorage)
	delChannel := make(chan model.BatchDeleteEntry, 3)
	uh := server.New(conf, short, delChannel)
	r := SetUpRouter(conf, uh)

	srv := httptest.NewServer(r)
	defer srv.Close()
	tSrv := newTestConf(srv, tStorage)

	res, err := json.Marshal(server.NewResult(conf.BaseURL() + "/" + "EEEEEEEE"))
	if err != nil {
		logger.Log.Error("marshal json", zap.Error(err))
	}
	tests := []test{
		{
			name:   "simple ShortenPost test #1",
			isMock: true,
			mockOn: func(m *mockedStorage) *mock.Call {
				return m.On("SaveURL", mock.Anything, mock.Anything,
					mock.Anything).Return("EEEEEEEE", nil)
			},
			reqFunc: func() *http.Request {
				url := server.NewURL("https://practicum.yandex.ru/")
				marshal, err := json.Marshal(url)
				if err != nil {
					logger.Log.Error("marshal json", zap.Error(err))
				}
				req := httptest.NewRequest("POST", srv.URL+"/api/shorten",
					bytes.NewReader(marshal))
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType: server.ApplicationJSON,
				statusCode:  201,
				body:        string(res),
			},
		},
		{
			name:   "empty body url ShortenPost test #2",
			isMock: false,
			reqFunc: func() *http.Request {
				url := server.NewURL("")
				marshal, err := json.Marshal(url)
				if err != nil {
					logger.Log.Error("marshal json", zap.Error(err))
				}
				req := httptest.NewRequest("POST", srv.URL+"/api/shorten",
					bytes.NewReader(marshal))
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType: server.TextPlain,
				statusCode:  400,
			},
		},
		{
			name:   "empty json ShortenPost test #3",
			isMock: false,
			reqFunc: func() *http.Request {
				body := strings.NewReader("{}")
				req := httptest.NewRequest("POST", srv.URL+"/api/shorten", body)
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType: server.TextPlain,
				statusCode:  400,
			},
		},
	}
	RunSubTests(t, tests, tSrv)
}

func TestShortenBatch(t *testing.T) {
	conf := config.Get()
	tStorage := new(mockedStorage)
	short := shortener.New(tStorage)
	delChannel := make(chan model.BatchDeleteEntry, 3)
	uh := server.New(conf, short, delChannel)
	r := SetUpRouter(conf, uh)

	srv := httptest.NewServer(r)
	defer srv.Close()
	tSrv := newTestConf(srv, tStorage)

	bResp := make([]model.BatchRespEntry, 2)
	b0 := model.NewBatchRespEntry("123", conf.BaseURL()+"/"+"EEEEEEEE")
	b1 := model.NewBatchRespEntry("321", conf.BaseURL()+"/"+"JJJJJJJJ")
	bResp[0] = b0
	bResp[1] = b1
	exp, err := json.Marshal(bResp)

	if err != nil {
		logger.Log.Error("marshal json", zap.Error(err))
	}
	tests := []test{
		{
			name:   "simple ShortenBatch test #1",
			isMock: true,
			mockOn: func(m *mockedStorage) *mock.Call {
				return m.On("SaveURLBatch", mock.Anything, mock.Anything,
					mock.Anything).Return(bResp, nil)
			},
			reqFunc: func() *http.Request {
				bReq := make([]model.BatchReqEntry, 2)
				entry0 := model.NewBatchReqEntry("123", "https://practicum.yandex.ru/")
				entry1 := model.NewBatchReqEntry("321", "https://yandex.ru/")
				bReq[0] = entry0
				bReq[1] = entry1
				marshal, err := json.Marshal(bReq)
				if err != nil {
					logger.Log.Error("marshal json", zap.Error(err))
				}
				req := httptest.NewRequest("POST", srv.URL+"/api/shorten/batch", bytes.NewReader(marshal))
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType: server.ApplicationJSON,
				statusCode:  201,
				body:        string(exp),
			},
		},
		{
			name:   "empty json ShortenBatch test #2",
			isMock: false,
			reqFunc: func() *http.Request {
				body := strings.NewReader("{}")
				req := httptest.NewRequest("POST", srv.URL+"/api/shorten/batch", body)
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType: server.TextPlain,
				statusCode:  400,
			},
		},
	}
	RunSubTests(t, tests, tSrv)
}

func TestNoRoutes(t *testing.T) {
	conf := config.Get()
	tStorage := new(mockedStorage)
	short := shortener.New(tStorage)
	delChannel := make(chan model.BatchDeleteEntry, 3)
	uh := server.New(conf, short, delChannel)
	r := SetUpRouter(conf, uh)

	srv := httptest.NewServer(r)
	defer srv.Close()
	tSrv := newTestConf(srv, tStorage)

	tests := []test{
		{
			name:   "bad url Get test '/test/AAAAAAAA' #1",
			isMock: false,
			reqFunc: func() *http.Request {
				req := httptest.NewRequest("GET", tSrv.URL+conf.BasePath()+"/test/AAAAAAAA", nil)
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType: server.TextPlain,
				statusCode:  400,
			},
		},
		{
			name:   "bad url Post test '/test' #2",
			isMock: false,
			reqFunc: func() *http.Request {
				body := strings.NewReader("https://practicum.yandex.ru/")
				req := httptest.NewRequest("POST", tSrv.URL+"/test", body)
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType: server.TextPlain,
				statusCode:  400,
			},
		},
		{
			name:   "bad method test 'DELETE' #3",
			isMock: false,
			reqFunc: func() *http.Request {
				body := strings.NewReader("https://practicum.yandex.ru/")
				req := httptest.NewRequest("DELETE", tSrv.URL+"/", body)
				req.RequestURI = ""
				return req
			},
			want: want{
				contentType: server.TextPlain,
				statusCode:  400,
			},
		},
	}
	RunSubTests(t, tests, tSrv)
}

func TestGzipCompression(t *testing.T) {
	conf := config.Get()
	tStorage := new(mockedStorage)
	tStorage.On("SaveURL", mock.Anything, mock.Anything,
		mock.Anything).Return("MMMMMMMM", nil)
	short := shortener.New(tStorage)
	delChannel := make(chan model.BatchDeleteEntry, 3)
	uh := server.New(conf, short, delChannel)
	r := SetUpRouter(conf, uh)

	srv := httptest.NewServer(r)
	defer srv.Close()

	requestBody := `{
    	"url": "https://practicum.yandex.ru/"
	}`

	successBody := `{
        "result":"` + conf.BaseURL() + "/MMMMMMMM" + `"
	}`
	client := createHTTPAuthClient(t, srv)
	t.Run("sends_gzip_json", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL+"/api/shorten", buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")

		resp, err := client.Do(r)
		require.NoError(t, err)

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip_json", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL+"/api/shorten", buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := client.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
	t.Run("accepts_gzip_text", func(t *testing.T) {
		buf := bytes.NewBufferString("https://practicum.yandex.ru/")
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := client.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.True(t, IDURLRegex.MatchString(string(b)))
	})
}
