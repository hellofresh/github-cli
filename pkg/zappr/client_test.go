package zappr

import (
	"net/http"
	"testing"

	"github.com/hashicorp/errwrap"
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

func TestArbitraryErrorWithJSONErrorBodyFromZappr(t *testing.T) {
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
	zapprMock.On("Handle", "DELETE", "/api/repos/1/approval", mock.Anything, mock.Anything).Return(test.Response{
		Status: http.StatusServiceUnavailable,
		Body: []byte(`{
			"type": "error",
			"status": 503,
			"detail": "Strange error message",
			"title": "Unplanned error"
		}`),
	})

	err := client.Disable(1)

	// Assert "unknown zappr" error was returned
	assert.True(t, errwrap.Contains(err, ErrZapprServerError.Error()))

	// Assert error detail retured by zappr is in returned error
	assert.True(t, errwrap.Contains(err, "Strange error message"), err)

	// Assert expected calls we made to Zappr
	zapprMock.AssertExpectations(t)
}

func TestArbitraryErrorWithNonJSONErrorBodyFromZappr(t *testing.T) {
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
	zapprMock.On("Handle", "DELETE", "/api/repos/1/approval", mock.Anything, mock.Anything).Return(test.Response{
		Status: http.StatusServiceUnavailable,
		Body:   []byte(`Strange error message`),
	})

	err := client.Disable(1)

	// Assert "unknown zappr" error was returned
	assert.True(t, errwrap.Contains(err, ErrZapprServerError.Error()))

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

	// Assert "already enabled" error was returned
	assert.Equal(t, ErrZapprAlreadyEnabled, err)

	// Assert expected calls we made to Zappr
	zapprMock.AssertExpectations(t)
}

func TestDisable(t *testing.T) {
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
	zapprMock.On("Handle", "DELETE", "/api/repos/1/approval", mock.Anything, mock.Anything).Return(test.Response{
		Status: http.StatusNoContent,
	})

	err := client.Disable(1)

	// Assert no errors were received
	assert.Nil(t, err)

	// Assert expected calls we made to Zappr
	zapprMock.AssertExpectations(t)
}

func TestDisable_RepoDeletedFromGithub(t *testing.T) {
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
	zapprMock.On("Handle", "DELETE", "/api/repos/1/approval", mock.Anything, mock.Anything).Return(test.Response{
		Status: http.StatusServiceUnavailable,
		Body: []byte(`{
			"type": "github error",
			"status": 503,
			"detail": "GET https://api.github.com/repos/myorg/myrepo/branches/master/protection/required_status_checks 404 Not Found",
			"title": "Github API Error"
		}`),
	})

	err := client.Disable(1)

	// Assert "already not enabled" error was returned
	assert.Equal(t, ErrZapprAlreadyNotEnabled, err)

	// Assert expected calls we made to Zappr
	zapprMock.AssertExpectations(t)
}

func TestDisable_RepoDeletedFromGithub_AndNotOnZappr(t *testing.T) {
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
	zapprMock.On("Handle", "DELETE", "/api/repos/1/approval", mock.Anything, mock.Anything).Return(test.Response{
		Status: http.StatusServiceUnavailable,
		Body: []byte(`{
			"type": "repository error",
			"status": 503,
			"detail": "Repository 1 not found.",
			"title": "Error during repository handling"
		}`),
	})

	err := client.Disable(1)

	// Assert "already not enabled" error was returned
	assert.Equal(t, ErrZapprAlreadyNotEnabled, err)

	// Assert expected calls we made to Zappr
	zapprMock.AssertExpectations(t)
}
