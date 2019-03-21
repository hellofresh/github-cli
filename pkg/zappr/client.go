package zappr

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dghubble/sling"
	"github.com/hashicorp/errwrap"
)

// Client represents a zappr client
type Client interface {
	Enable(repoID int) error
}

type clientImpl struct {
	zapprToken  string
	githubToken string
	slingClient *sling.Sling
}

type zapprErrorResponse struct {
	Type   string `json:"type,omitempty"`
	Detail string `json:"detail,omitempty"`
	Title  string `json:"title,omitempty"`
}

var (
	timeout = 30 * time.Second

	// ErrZapprUnauthorized is used when we receive a 401 from Zappr
	ErrZapprUnauthorized = errors.New("you do not have permissions to use zappr api")

	// ErrZapprAlreadyExist is used when "Enabled" is called for a repo that Zappr is already enabled for
	ErrZapprAlreadyExist = errors.New("zappr already enabled for the repo")

	// ErrZapprServerError is used when we receive a code different than 200 from Zappr
	ErrZapprServerError = errors.New("unknown error from zappr")
)

// NewWithZapprToken creates a new Zappr client that uses Zappr Token to make calls to Zappr
func NewWithZapprToken(zapprURL string, zapprToken string, httpClient *http.Client) Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	slingClient := sling.New().Client(httpClient).Base(zapprURL)

	return &clientImpl{
		zapprToken:  zapprToken,
		slingClient: slingClient,
	}
}

// NewWithGithubToken creates a new Zappr client that uses Github Token to make calls to Zappr
func NewWithGithubToken(zapprURL string, githubToken string, httpClient *http.Client) Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	slingClient := sling.New().Client(httpClient).Base(zapprURL)

	return &clientImpl{
		githubToken: githubToken,
		slingClient: slingClient,
	}
}

// Enable turns on Zappr approval check on a Github repo
func (c *clientImpl) Enable(repoID int) error {
	req, err := c.slingClient.Get(fmt.Sprintf("api/repos/%d?autoSync=true", repoID)).Request()
	if err != nil {
		return errwrap.Wrapf("could not fetch repo on zappr to enable approval check: {{err}}", err)
	}

	err = c.doRequest(req)
	if err != nil {
		return errwrap.Wrapf("could not fetch repo on zappr to enable approval check: {{err}}", err)
	}

	req, err = c.slingClient.Put(fmt.Sprintf("%d/approval", repoID)).Request()
	if err != nil {
		return errwrap.Wrapf("could not Enable Zappr approval checks on repo: {{err}}", err)
	}

	return c.doRequest(req)
}

func (c *clientImpl) doRequest(req *http.Request) error {
	if c.zapprToken == "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", c.githubToken))
	} else {
		req.Header.Add("Cookie", c.zapprToken)
	}

	zapprErrorResponse := &zapprErrorResponse{}
	resp, err := c.slingClient.Do(req, nil, zapprErrorResponse)

	// Not checking if err is nil here because this err object is related to decoding
	// the response body to the zapprErrorResponse object and not the http call itself.
	// The zapprErrorResponse object is only needed to extract the error message when
	// Zappr returns a proper error message in JSON format,
	// if not (JSON decode would fail) move on and use the HTTP status code

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrZapprUnauthorized
	}

	if resp.StatusCode == http.StatusServiceUnavailable && err == nil && strings.HasPrefix(zapprErrorResponse.Detail, "Check approval already exists for repository") {
		return ErrZapprAlreadyExist
	}

	if resp.StatusCode >= 400 {
		if err == nil {
			return errors.New(zapprErrorResponse.Detail)
		}

		return ErrZapprServerError
	}

	return nil
}
