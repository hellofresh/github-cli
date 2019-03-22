package zappr

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/hellofresh/github-cli/pkg/test"
)

func getDefaultHeader() *http.Header {
	header := &http.Header{}
	header.Add("User-Agent", "Go-http-client/1.1")
	header.Add("Accept-Encoding", "gzip")
	header.Add("Content-Length", "0")

	return header
}

func getGithubAuthHeader(token string) *http.Header {
	authHeader := getDefaultHeader()
	authHeader.Add("Authorization", fmt.Sprintf("token %s", token))

	return authHeader
}

func getZapprAuthHeader(token string) *http.Header {
	authHeader := getDefaultHeader()
	authHeader.Add("Cookie", token)

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

// NewMockAndHandler returns a Zappr Client that uses Github Token, Mockhandler, and Server. The client proxies
// requests to the server and handlers can be registered on the mux to handle
// requests. The caller must close the test server.
func NewMockAndHandler() (Client, *test.MockHandler, *test.Server) {
	return NewMockAndHandlerWithGithubToken("1234567890")
}

// NewMockAndHandlerWithGithubToken returns a Zappr Client that uses Github Token, Mockhandler, and Server. The client proxies
// requests to the server and handlers can be registered on the mux to handle
// requests. The caller must close the test server.
func NewMockAndHandlerWithGithubToken(githubToken string) (Client, *test.MockHandler, *test.Server) {
	httpClient, mockHandler, mockServer := newMockAndHandler()
	client := NewWithGithubToken("https://fake.zappr/", githubToken, httpClient)

	return client, mockHandler, mockServer
}

// NewMockAndHandlerWithZapprToken returns a Zappr Client that uses Zappr Token, Mockhandler, and Server. The client proxies
// requests to the server and handlers can be registered on the mux to handle
// requests. The caller must close the test server.
func NewMockAndHandlerWithZapprToken(zapprToken string) (Client, *test.MockHandler, *test.Server) {
	httpClient, mockHandler, mockServer := newMockAndHandler()
	client := NewWithZapprToken("https://fake.zappr/", zapprToken, httpClient)

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
