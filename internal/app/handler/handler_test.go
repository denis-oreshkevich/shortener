package handler

import (
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPost(t *testing.T) {
	postTests := []Test{
		{
			name:   "simple post test #1",
			isMock: true,
			mockOn: func(m *MockedRepository) *mock.Call {
				return m.On("SaveURL", mock.Anything, mock.Anything).Return()
			},
			reqFunc: func() *http.Request {
				body := strings.NewReader("https://practicum.yandex.ru/")
				return httptest.NewRequest("POST", "/", body)
			},
			want: Want{
				contentType: "text/plain",
				statusCode:  201,
			},
		},
		{
			name:   "nil body post test #2",
			isMock: false,
			reqFunc: func() *http.Request {
				return httptest.NewRequest("POST", "/", nil)
			},
			want: Want{
				contentType: "text/plain",
				statusCode:  400,
			},
		},
		{
			name:   "bad body url post test #3",
			isMock: false,
			reqFunc: func() *http.Request {
				body := strings.NewReader("ahahahah")
				return httptest.NewRequest("POST", "/", body)
			},
			want: Want{
				contentType: "text/plain",
				statusCode:  400,
			},
		},
	}
	RunSubTests(t, postTests)
}

func TestGet(t *testing.T) {
	getTests := []Test{
		{
			name:   "simple get test #1",
			isMock: true,
			mockOn: func(m *MockedRepository) *mock.Call {
				return m.On("FindURL", "AAAAAAAA").Return("http://localhost:30001/", true)
			},
			reqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/AAAAAAAA", nil)
			},
			want: Want{
				contentType:    "text/plain",
				statusCode:     307,
				headerLocation: "http://localhost:30001/",
			},
		},
		{
			name:   "bad id get test #2",
			isMock: false,
			reqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/HHH", nil)
			},
			want: Want{
				contentType: "text/plain",
				statusCode:  400,
			},
		},
		{
			name:   "not stored url get test #3",
			isMock: true,
			mockOn: func(m *MockedRepository) *mock.Call {
				return m.On("FindURL", "HHHHHHHH").Return("", false)
			},
			reqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/HHHHHHHH", nil)
			},
			want: Want{
				contentType: "text/plain",
				statusCode:  400,
			},
		},
	}
	RunSubTests(t, getTests)
}

func TestNoRoutes(t *testing.T) {
	tests := []Test{
		{
			name:   "bad url get test '/test/AAAAAAAA' #1",
			isMock: false,
			reqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/test/AAAAAAAA", nil)
			},
			want: Want{
				contentType: "text/plain",
				statusCode:  400,
			},
		},
		{
			name:   "bad url post test '/test' #2",
			isMock: false,
			reqFunc: func() *http.Request {
				body := strings.NewReader("https://practicum.yandex.ru/")
				return httptest.NewRequest("POST", "/test", body)
			},
			want: Want{
				contentType: "text/plain",
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
				contentType: "text/plain",
				statusCode:  400,
			},
		},
	}
	RunSubTests(t, tests)
}
