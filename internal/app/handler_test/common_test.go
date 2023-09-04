package handler

import (
	"github.com/denis-oreshkevich/shortener/internal/app/constant"
	"github.com/stretchr/testify/mock"
	"net/http"
	"regexp"
)

var IdRegex = regexp.MustCompile(constant.ServerURL + "/[A-Za-z]{8}")

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
