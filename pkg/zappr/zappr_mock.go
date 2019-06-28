package zappr

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hellofresh/github-cli/pkg/test"
)

func getDefaultHeader(addContentLength bool) *http.Header {
	header := &http.Header{}
	header.Add("User-Agent", "Go-http-client/1.1")
	header.Add("Accept-Encoding", "gzip")

	if addContentLength {
		header.Add("Content-Length", "0")
	}

	return header
}

func getGithubAuthHeader(token string, addContentLength bool) *http.Header {
	authHeader := getDefaultHeader(addContentLength)
	authHeader.Add("Authorization", fmt.Sprintf("token %s", token))

	return authHeader
}

func newMockAndHandler() (*http.Client, *test.MockHandler, *test.Server) {
	mockHandler := &test.MockHandler{}
	mockServer := test.NewUnstartedServer(mockHandler)

	transport := &RewriteTransport{&http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(mockServer.URL())
		},
	}}

	httpClient := &http.Client{Transport: transport}

	return httpClient, mockHandler, mockServer
}

func newMockAndHandlerWithNilResponseTransport() (*http.Client, *test.MockHandler, *test.Server) {
	mockHandler := &test.MockHandler{}
	mockServer := test.NewUnstartedServer(mockHandler)

	httpClient := &http.Client{
		Transport: &NilResponseTransport{},
	}

	return httpClient, mockHandler, mockServer
}

// NewMockAndHandler returns a Zappr Client that uses Github Token, Mockhandler, and Server. The client proxies
// requests to the server and handlers can be registered on the mux to handle
// requests. The caller must close the test server.
func NewMockAndHandler() (Client, *test.MockHandler, *test.Server) {
	return NewMockAndHandlerWithGithubToken("1234567890")
}

// NewMockAndHandlerNilResponse returns a Zappr Client that uses Github Token, Mockhandler, and Server.
// The internal http client always returns a nil response object.
func NewMockAndHandlerNilResponse() (Client, *test.MockHandler, *test.Server) {
	httpClient, mockHandler, mockServer := newMockAndHandlerWithNilResponseTransport()
	client := New("https://fake.zappr/", "1234567890", httpClient)

	return client, mockHandler, mockServer
}

// NewMockAndHandlerWithGithubToken returns a Zappr Client that uses Github Token, Mockhandler, and Server. The client proxies
// requests to the server and handlers can be registered on the mux to handle
// requests. The caller must close the test server.
func NewMockAndHandlerWithGithubToken(githubToken string) (Client, *test.MockHandler, *test.Server) {
	httpClient, mockHandler, mockServer := newMockAndHandler()
	client := New("https://fake.zappr", githubToken, httpClient)

	return client, mockHandler, mockServer
}

// RewriteTransport rewrites https requests to http to avoid TLS cert issues
// during testing.
type RewriteTransport struct {
	Transport http.RoundTripper
}

// RoundTrip rewrites the request scheme to http and calls through to the
// composed RoundTripper or if it is nil, to the http.DefaultTransport.
func (t *RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	if t.Transport == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return t.Transport.RoundTrip(req)
}

// NilResponseTransport always returns a nil response object
type NilResponseTransport struct {
	Transport http.RoundTripper
}

// RoundTrip rewrites the request scheme to http and calls through to the
// composed RoundTripper or if it is nil, to the http.DefaultTransport.
func (t *NilResponseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("some error")
}
