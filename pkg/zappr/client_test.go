package zappr

import (
	"net/http"
	"testing"

	"github.com/hellofresh/github-cli/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthWithGithubToken(t *testing.T) {
	token := "abcdefgh"

	// Get the Zappr Client, Mock Handler and Test Server
	client, zapprMock, testServer := NewMockAndHandlerWithGithubToken(token)

	// Start the test server and stop it when done
	testServer.Start()
	defer testServer.Close()

	// Should call the "fetch repo" zappr endpoint to find the just created repo
	zapprMock.On("Handle", "GET", "/api/repos/1?autoSync=true", mock.Anything, mock.Anything).Return(test.Response{
		Status: http.StatusOK,
	})

	// Add handlers to define expected call(s) to, and response(s) from Zappr
	zapprMock.On("Handle", "PUT", "/api/repos/1/approval", getGithubAuthHeader(token), mock.Anything).Return(test.Response{
		Status: http.StatusCreated,
	})

	client.Enable(1)

	// Assert expected calls we made to Zappr
	zapprMock.AssertExpectations(t)
}

func TestAuthWithZapprToken(t *testing.T) {
	token := "abcdefgh"

	// Get the Zappr Client, Mock Handler and Test Server
	client, zapprMock, testServer := NewMockAndHandlerWithZapprToken(token)

	// Start the test server and stop it when done
	testServer.Start()
	defer testServer.Close()

	// Should call the "fetch repo" zappr endpoint to find the just created repo
	zapprMock.On("Handle", "GET", "/api/repos/1?autoSync=true", mock.Anything, mock.Anything).Return(test.Response{
		Status: http.StatusOK,
	})

	// Add handlers to define expected call(s) to, and response(s) from Zappr
	zapprMock.On("Handle", "PUT", "/api/repos/1/approval", getZapprAuthHeader(token), mock.Anything).Return(test.Response{
		Status: http.StatusCreated,
	})

	client.Enable(1)

	// Assert expected calls we made to Zappr
	zapprMock.AssertExpectations(t)
}

func TestEnable(t *testing.T) {
	// Get the Zappr Client, Mock Handler and Test Server
	client, zapprMock, testServer := NewMockAndHandler()

	// Start the test server and stop it when done
	testServer.Start()
	defer testServer.Close()

	// Should call the "fetch repo" zappr endpoint to find the just created repo
	zapprMock.On("Handle", "GET", "/api/repos/1?autoSync=true", mock.Anything, mock.Anything).Return(test.Response{
		Status: http.StatusOK,
	})

	// Should call the "put approval" endpoint to enable zappr on the repo
	zapprMock.On("Handle", "PUT", "/api/repos/1/approval", mock.Anything, mock.Anything).Return(test.Response{
		Status: http.StatusCreated,
	})

	err := client.Enable(1)

	// Assert no errors were received
	assert.Nil(t, err)

	// Assert expected calls we made to Zappr
	zapprMock.AssertExpectations(t)
}

func TestEnable_AlreadyExist(t *testing.T) {
	// Get the Zappr Client, Mock Handler and Test Server
	client, zapprMock, testServer := NewMockAndHandler()

	// Start the test server and stop it when done
	testServer.Start()
	defer testServer.Close()

	// Should call the "fetch repo" zappr endpoint to find the just created repo
	zapprMock.On("Handle", "GET", "/api/repos/1?autoSync=true", mock.Anything, mock.Anything).Return(test.Response{
		Status: http.StatusOK,
	})

	// Should call the "put approval" endpoint to enable zappr on the repo
	zapprMock.On("Handle", "PUT", "/api/repos/1/approval", mock.Anything, mock.Anything).Return(test.Response{
		Status: http.StatusServiceUnavailable,
		Body: []byte(`{
			"type": "check error",
			"status": 503,
			"detail": "Check approval already exists for repository 1.",
			"title": "Error during check processing"
		}`),
	})

	err := client.Enable(1)

	// Assert "already exist" error was returned
	assert.Equal(t, ErrZapprAlreadyExist, err)

	// Assert expected calls we made to Zappr
	zapprMock.AssertExpectations(t)
}
