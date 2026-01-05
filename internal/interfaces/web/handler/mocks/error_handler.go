package mocks

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockErrorHandler struct {
	mock.Mock
}

func (m *MockErrorHandler) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	m.Called(w, r, err)
}

func (m *MockErrorHandler) Error(w http.ResponseWriter, r *http.Request, status int, err error) {
	m.Called(w, r, status, err)
}

func (m *MockErrorHandler) LogServerError(r *http.Request, err error) {
	m.Called(r, err)
}
