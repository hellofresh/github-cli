package test

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockHandler is a test.Handler that uses github.com/stretchr/testify/mock.
type MockHandler struct {
	mock.Mock
}

// Handle makes this implement the Handler interface.
func (m *MockHandler) Handle(method, path string, header *http.Header, body []byte) Response {
	args := m.Called(method, path, header, body)
	return args.Get(0).(Response)
}
