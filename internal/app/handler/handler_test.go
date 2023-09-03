package handler

import (
	"github.com/denis-oreshkevich/shortener/internal/app/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

var regex = regexp.MustCompile(constant.ServerURL + "/[A-Za-z]{8}")

type want struct {
	contentType    string
	statusCode     int
	body           string
	headerLocation string
}
type Req struct {
	url    string
	method string
	body   string
}
type test struct {
	name   string
	isMock bool
	mockOn func(m *mockedRepository) *mock.Call
	req    Req
	want   want
}

type mockedRepository struct {
	mock.Mock
}

func (m *mockedRepository) SaveURL(id, url string) {
	m.Called(id, url)
}

func (m *mockedRepository) FindURL(id string) (string, bool) {
	args := m.Called(id)
	return args.String(0), args.Bool(1)
}

func TestHandlePost(t *testing.T) {
	postTests := []test{
		{
			name:   "simple post test #1",
			isMock: true,
			mockOn: func(m *mockedRepository) *mock.Call {
				return m.On("SaveURL", mock.Anything, mock.Anything).Return()
			},
			req: Req{
				url:    "/",
				method: "POST",
				body:   "https://practicum.yandex.ru/",
			},
			want: want{
				contentType: "text/plain",
				statusCode:  201,
			},
		},
		{
			name:   "bad url post test #2",
			isMock: false,
			req: Req{
				url:    "/test",
				method: "POST",
				body:   "https://practicum.yandex.ru/",
			},
			want: want{
				contentType: "text/plain",
				statusCode:  400,
			},
		},
		{
			name:   "bad body url post test #3",
			isMock: false,
			req: Req{
				url:    "/",
				method: "POST",
				body:   "ahahahah",
			},
			want: want{
				contentType: "text/plain",
				statusCode:  400,
			},
		},
	}
	runSubTests(t, postTests)
}

func TestHandleGet(t *testing.T) {
	getTests := []test{
		{
			name:   "simple get test #1",
			isMock: true,
			mockOn: func(m *mockedRepository) *mock.Call {
				return m.On("FindURL", "/AAAAAAAA").Return("http://localhost:30001/", true)
			},
			req: Req{
				url:    "/AAAAAAAA",
				method: "GET",
			},
			want: want{
				contentType:    "text/plain",
				statusCode:     307,
				headerLocation: "http://localhost:30001/",
			},
		},
		{
			name:   "bad url get test #2",
			isMock: false,
			req: Req{
				url:    "/test/AAAAAAAA",
				method: "GET",
				body:   "https://practicum.yandex.ru/",
			},
			want: want{
				contentType: "text/plain",
				statusCode:  400,
			},
		},
		{
			name:   "not stored url get test #3",
			isMock: true,
			mockOn: func(m *mockedRepository) *mock.Call {
				return m.On("FindURL", "/HHHHHHHH").Return("", false)
			},
			req: Req{
				url:    "/HHHHHHHH",
				method: "GET",
			},
			want: want{
				contentType: "text/plain",
				statusCode:  400,
			},
		},
	}
	runSubTests(t, getTests)
}

func runSubTests(t *testing.T, tests []test) {
	testObj := new(mockedRepository)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request *http.Request
			if tt.req.body != "" {
				body := strings.NewReader(tt.req.body)
				request = httptest.NewRequest(tt.req.method, tt.req.url, body)
			} else {
				request = httptest.NewRequest(tt.req.method, tt.req.url, nil)

			}
			w := httptest.NewRecorder()

			var mockCall *mock.Call

			if tt.isMock {
				mockCall = tt.mockOn(testObj)
			}

			handlerFunc := URL(testObj)
			handlerFunc.ServeHTTP(w, request)

			if tt.isMock {
				testObj.AssertExpectations(t)
				mockCall.Unset()
			}
			result := w.Result()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			if tt.want.headerLocation != "" {
				assert.Equal(t, tt.want.headerLocation, result.Header.Get("Location"))
			}
			if tt.want.body != "" {
				defer result.Body.Close()
				respBody, err := io.ReadAll(result.Body)
				assert.NoError(t, err)
				assert.True(t, regex.MatchString(string(respBody)))
			}
		})
	}
}
