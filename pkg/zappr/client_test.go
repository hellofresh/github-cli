package zappr

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hellofresh/github-cli/pkg/test"
)

func TestAuthWithGithubToken(t *testing.T) {
	token := "abcdefgh"

	// Get the Zappr Client, Mock Handler and Test Server
	client, zapprMock, testServer := NewMockAndHandlerWithGithubToken(token)

	// Start the test server and stop it when done
	testServer.Start()
	defer testServer.Close()

	// Should call the "fetch repo" zappr endpoint to find the just created repo
	zapprMock.On("Handle", "GET", "/api/repos/1?autoSync=true", getGithubAuthHeader(token, false), mock.Anything).Return(test.Response{
		Status: http.StatusOK,
	})

	// Add handlers to define expected call(s) to, and response(s) from Zappr
	zapprMock.On("Handle", "PUT", "/api/repos/1/approval", getGithubAuthHeader(token, true), mock.Anything).Return(test.Response{
		Status: http.StatusCreated,
	})

	client.Enable(1)

	// Assert expected calls were made to Zappr
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
	require.Error(t, err)

	// Assert "unknown zappr" error was returned
	assert.True(t, strings.Contains(err.Error(), ErrZapprServerError.Error()))

	// Assert error detail retured by zappr is in returned error
	assert.True(t, strings.Contains(err.Error(), "Strange error message"))

	// Assert expected calls were made to Zappr
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
	require.Error(t, err)

	// Assert "unknown zappr" error was returned
	assert.True(t, strings.Contains(err.Error(), ErrZapprServerError.Error()))

	// Assert expected calls were made to Zappr
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
	assert.NoError(t, err)

	// Assert expected calls were made to Zappr
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
	require.Error(t, err)

	// Assert "already enabled" error was returned
	assert.True(t, errors.Is(err, ErrZapprAlreadyEnabled))

	// Assert expected calls were made to Zappr
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
	assert.NoError(t, err)

	// Assert expected calls were made to Zappr
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
	require.Error(t, err)

	// Assert "already not enabled" error was returned
	assert.Equal(t, ErrZapprAlreadyNotEnabled, err)

	// Assert expected calls were made to Zappr
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
	require.Error(t, err)

	// Assert "already not enabled" error was returned
	assert.Equal(t, ErrZapprAlreadyNotEnabled, err)

	// Assert expected calls were made to Zappr
	zapprMock.AssertExpectations(t)
}

func TestProblematicRequest(t *testing.T) {
	// Get the Zappr Client, Mock Handler and Test Server
	client, zapprMock, testServer := NewMockAndHandlerNilResponse()

	// Start the test server and stop it when done
	testServer.Start()
	defer testServer.Close()

	err := client.Enable(1)
	require.Error(t, err)

	// Assert "unknown zappr" error was returned
	assert.True(t, strings.Contains(err.Error(), ErrZapprServerError.Error()))

	// Assert expected calls were made to Zappr
	zapprMock.AssertExpectations(t)
}

func TestImpersonateGitHubApp(t *testing.T) {
	token := "12345678"
	zapprAppToken := "abcdefgh"

	// Get the Zappr Client, Mock Handler and Test Server
	client, zapprMock, testServer := NewMockAndHandlerWithGithubToken(token)

	// Start the test server and stop it when done
	testServer.Start()
	defer testServer.Close()

	// Should call the "apptoken" zappr endpoint to get the github token representing zappr app and use the users github token
	zapprMock.On("Handle", "GET", "/api/apptoken", getGithubAuthHeader(token, false), mock.Anything).Return(test.Response{
		Status: http.StatusOK,
		Body:   []byte(fmt.Sprintf(`{ "token": "%s" }`, zapprAppToken)),
	})

	err := client.ImpersonateGitHubApp()
	require.NoError(t, err)

	// Assert expected calls were made to Zappr
	zapprMock.AssertExpectations(t)

	// reset expectations, i need to use the same mock object that is using the retrieved zappr app github token
	zapprMock.ExpectedCalls = []*mock.Call{}

	// Should call the "fetch repo" zappr endpoint to find the just created repo and use zappr app's github token
	zapprMock.On("Handle", "GET", "/api/repos/1?autoSync=true", getGithubAuthHeader(zapprAppToken, false), mock.Anything).Return(test.Response{
		Status: http.StatusOK,
	})

	// Add handlers to define expected call(s) to, and response(s) from Zappr and use zappr app's github token
	zapprMock.On("Handle", "PUT", "/api/repos/1/approval", getGithubAuthHeader(zapprAppToken, true), mock.Anything).Return(test.Response{
		Status: http.StatusCreated,
	})

	client.Enable(1)

	// Assert expected calls were made to Zappr
	zapprMock.AssertExpectations(t)
}
