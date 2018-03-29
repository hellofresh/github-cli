package pullapprove

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/errwrap"
)

// Client represents a pull approve client
type Client struct {
	BaseURL string
	Token   string
}

var (
	// ErrPullApproveUnauthorized is used when we receive a 401 from pull approve
	ErrPullApproveUnauthorized = errors.New("you do not have permissions to use pull approve API")

	// ErrPullApproveServerError is used when we receive a code different than 200 from pull approve
	ErrPullApproveServerError = errors.New("could not create a pull approve repository")
)

// New creates a new instance of Client
func New(token string) *Client {
	return &Client{
		BaseURL: "https://pullapprove.com/api",
		Token:   token,
	}
}

// Create creates a new repository on pull approve
func (c *Client) Create(name string, organization string) error {
	body := []byte(fmt.Sprintf(`{"name": "%s"}`, name))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/orgs/%s/repos/", c.BaseURL, organization), bytes.NewBuffer(body))
	if err != nil {
		return errwrap.Wrapf("could not create the request to pull approve: {{err}}", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.Token))
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errwrap.Wrapf("could not create a pull approve repository: {{err}}", err)
	}

	if res.StatusCode == http.StatusUnauthorized {
		return ErrPullApproveUnauthorized
	}

	if res.StatusCode != http.StatusCreated {
		return ErrPullApproveServerError
	}

	return nil
}
