package handler

import (
	"github.com/denis-oreshkevich/shortener/internal/app/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

var IDURLRegex = regexp.MustCompile(constant.ServerURL + "/[A-Za-z]{8}")

type MockedRepository struct {
	mock.Mock
}

func (m *MockedRepository) SaveURL(id, url string) {
	m.Called(id, url)
}

func (m *MockedRepository) FindURL(id string) (string, bool) {
	args := m.Called(id)
	return args.String(0), args.Bool(1)
}

type Want struct {
	contentType    string
	statusCode     int
	body           string
	headerLocation string
}

type Test struct {
	name    string
	isMock  bool
	mockOn  func(m *MockedRepository) *mock.Call
	reqFunc func() *http.Request
	want    Want
}

func RunSubTests(t *testing.T, tests []Test) {
	testObj := new(MockedRepository)
	router := SetupRouter(testObj)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.reqFunc()
			w := httptest.NewRecorder()

			var mockCall *mock.Call

			if tt.isMock {
				mockCall = tt.mockOn(testObj)
			}

			router.ServeHTTP(w, request)

			if tt.isMock {
				testObj.AssertExpectations(t)
				mockCall.Unset()
			}
			result := w.Result()
			defer result.Body.Close()
			Assert(t, tt, result)
		})
	}
}

func Assert(t *testing.T, tt Test, result *http.Response) {
	assert.Equal(t, tt.want.statusCode, result.StatusCode)
	assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
	if tt.want.headerLocation != "" {
		assert.Equal(t, tt.want.headerLocation, result.Header.Get("Location"))
	}
	if tt.want.body != "" {
		respBody, err := io.ReadAll(result.Body)
		assert.NoError(t, err)
		assert.True(t, IDURLRegex.MatchString(string(respBody)))
	}
}
