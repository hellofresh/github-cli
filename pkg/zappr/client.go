package zappr

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dghubble/sling"
)

// Client represents a zappr client
type Client interface {
	Enable(repoID int) error
	Disable(repoID int) error
	ImpersonateGitHubApp() error // Retrieves the Github token representing Zappr and uses that for future requests
}

type clientImpl struct {
	githubToken         string
	githubZapprAppToken string
	slingClient         *sling.Sling
}

type zapprErrorResponse struct {
	Type   string `json:"type,omitempty"`
	Detail string `json:"detail,omitempty"`
	Title  string `json:"title,omitempty"`
}

type zapprAppTokenResponse struct {
	Token string `json:"token"`
}

var (
	timeout = 30 * time.Second

	// ErrZapprUnauthorized is used when we receive a 401 from Zappr
	ErrZapprUnauthorized = errors.New("you do not have permissions to use zappr api")

	// ErrZapprAlreadyEnabled is used when "Enable" is called for a repo that Zappr is already enabled for
	ErrZapprAlreadyEnabled = errors.New("zappr already enabled for the repo")

	// ErrZapprAlreadyNotEnabled is used when "Disable" is called for a repo that Zappr is already NOT enabled for
	ErrZapprAlreadyNotEnabled = errors.New("zappr is already not enabled for the repo")

	// ErrZapprServerError is used when we receive a code different than 200 from Zappr
	ErrZapprServerError = errors.New("unknown error from zappr")
)

// New creates a new Zappr client that uses Github Token to make calls to Zappr
func New(zapprURL string, githubToken string, httpClient *http.Client) Client {
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
	req, err := c.slingClient.Get(fmt.Sprintf("/api/repos/%d?autoSync=true", repoID)).Request()
	if err != nil {
		return fmt.Errorf("could not fetch repo on zappr to enable approval check: %w", err)
	}

	status, zapprErrorResponse, err := c.doRequest(req, nil)
	if err != nil {
		return fmt.Errorf("could not fetch repo on zappr to enable approval check: %w", err)
	}

	req, err = c.slingClient.Put(fmt.Sprintf("%d/approval", repoID)).Request()
	if err != nil {
		return fmt.Errorf("could not Enable Zappr approval checks on repo: %w", err)
	}

	status, zapprErrorResponse, err = c.doRequest(req, nil)
	if status == http.StatusServiceUnavailable && zapprErrorResponse != nil {
		// Zappr already active on the repo
		if strings.HasPrefix(zapprErrorResponse.Detail, "Check approval already exists for repository") {
			return ErrZapprAlreadyEnabled
		}
	}

	return err
}

// Disable turns off Zappr approval check on a Github repo
func (c *clientImpl) Disable(repoID int) error {
	req, err := c.slingClient.Get(fmt.Sprintf("/api/repos/%d?autoSync=true", repoID)).Request()
	if err != nil {
		return fmt.Errorf("could not fetch repo on zappr to enable approval check: %w", err)
	}

	status, zapprErrorResponse, err := c.doRequest(req, nil)
	if err != nil {
		return fmt.Errorf("could not fetch repo on zappr to enable approval check: %w", err)
	}

	req, err = c.slingClient.Delete(fmt.Sprintf("%d/approval", repoID)).Request()
	if err != nil {
		return fmt.Errorf("could not Disable Zappr approval checks on repo: %w", err)
	}

	status, zapprErrorResponse, err = c.doRequest(req, nil)
	if status == http.StatusServiceUnavailable && zapprErrorResponse != nil {
		// Zappr active on the repo, but repo has been deleted from github
		if strings.HasSuffix(zapprErrorResponse.Detail, "required_status_checks 404 Not Found") {
			return ErrZapprAlreadyNotEnabled
		}

		// Repo is not on Zappr (enabled/disabled), and is also not on github (deleted or was not created at all)
		if zapprErrorResponse.Detail == fmt.Sprintf("Repository %d not found.", repoID) {
			return ErrZapprAlreadyNotEnabled
		}
	}

	return err
}

func (c *clientImpl) ImpersonateGitHubApp() error {
	req, err := c.slingClient.Get("api/apptoken").Request()
	if err != nil {
		return fmt.Errorf("could not fetch github token for zappr github app: %w", err)
	}

	tokenResponse := &zapprAppTokenResponse{}
	_, _, err = c.doRequest(req, tokenResponse)
	if err != nil {
		return fmt.Errorf("could not fetch github token for zappr github app: %w", err)
	}

	c.githubZapprAppToken = tokenResponse.Token
	return nil
}

func (c *clientImpl) doRequest(req *http.Request, response interface{}) (int, *zapprErrorResponse, error) {
	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.githubToken))

	if c.githubZapprAppToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", c.githubZapprAppToken))
	}

	zapprErrorResponse := &zapprErrorResponse{}
	resp, err := c.slingClient.Do(req, response, zapprErrorResponse)

	if resp == nil {
		// Even though 0 does not seem like a valid http response code
		// The status would be ignored by any calling code anyway since a non-nil error is returned
		return 0, nil, fmt.Errorf("%s: %w", ErrZapprServerError.Error(), err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return resp.StatusCode, nil, ErrZapprUnauthorized
	}

	if resp.StatusCode >= 400 {
		if err == nil {
			return resp.StatusCode, zapprErrorResponse, fmt.Errorf("%s: %w", ErrZapprServerError.Error(), errors.New(zapprErrorResponse.Detail))
		}

		return resp.StatusCode, nil, fmt.Errorf("%s: %w", ErrZapprServerError.Error(), err)
	}

	return resp.StatusCode, nil, err
}
