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

	// ErrZapprAlreadyEnabled is used when "Enable" is called for a repo that Zappr is already enabled for
	ErrZapprAlreadyEnabled = errors.New("zappr already enabled for the repo")

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

	status, zapprErrorResponse, err := c.doRequest(req)
	if err != nil {
		return errwrap.Wrapf("could not fetch repo on zappr to enable approval check: {{err}}", err)
	}

	req, err = c.slingClient.Put(fmt.Sprintf("%d/approval", repoID)).Request()
	if err != nil {
		return errwrap.Wrapf("could not Enable Zappr approval checks on repo: {{err}}", err)
	}

	status, zapprErrorResponse, err = c.doRequest(req)
	if status == http.StatusServiceUnavailable && zapprErrorResponse != nil {
		// Zappr already active on the repo
		if strings.HasPrefix(zapprErrorResponse.Detail, "Check approval already exists for repository") {
			return ErrZapprAlreadyEnabled
		}
	}

	return err
}

func (c *clientImpl) doRequest(req *http.Request) (int, *zapprErrorResponse, error) {
	if c.zapprToken == "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", c.githubToken))
	} else {
		req.Header.Add("Cookie", c.zapprToken)
	}

	zapprErrorResponse := &zapprErrorResponse{}
	resp, err := c.slingClient.Do(req, nil, zapprErrorResponse)

	if resp.StatusCode == http.StatusUnauthorized {
		return resp.StatusCode, nil, ErrZapprUnauthorized
	}

	if resp.StatusCode >= 400 {
		if err == nil {
			return resp.StatusCode, zapprErrorResponse, errwrap.Wrap(errors.New(zapprErrorResponse.Detail), ErrZapprServerError)
		}

		return resp.StatusCode, nil, ErrZapprServerError
	}

	return resp.StatusCode, nil, err
}
