package handler

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlePost(t *testing.T) {
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
			name:   "bad url post test #2",
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
	runSubTests(t, postTests)
}

func TestHandleGet(t *testing.T) {
	getTests := []Test{
		{
			name:   "simple get test #1",
			isMock: true,
			mockOn: func(m *MockedRepository) *mock.Call {
				return m.On("FindURL", "/AAAAAAAA").Return("http://localhost:30001/", true)
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
			name:   "bad url get test #2",
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
			name:   "not stored url get test #3",
			isMock: true,
			mockOn: func(m *MockedRepository) *mock.Call {
				return m.On("FindURL", "/HHHHHHHH").Return("", false)
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
	runSubTests(t, getTests)
}

func runSubTests(t *testing.T, tests []Test) {
	testObj := new(MockedRepository)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.reqFunc()
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
				assert.True(t, IDURLRegex.MatchString(string(respBody)))
			}
		})
	}
}
