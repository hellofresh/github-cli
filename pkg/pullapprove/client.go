package pullapprove

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/hashicorp/errwrap"
)

type (
	// Client represents a pull approve client
	Client struct {
		Token  string
		client *sling.Sling
	}

	createReq struct {
		Name string `json:"name"`
	}
)

var (
	// ErrPullApproveUnauthorized is used when we receive a 401 from pull approve
	ErrPullApproveUnauthorized = errors.New("you do not have permissions to use pull approve API")

	// ErrPullApproveServerError is used when we receive a code different than 200 from pull approve
	ErrPullApproveServerError = errors.New("could not create a pull approve repository")
)

// New creates a new instance of Client
func New(token string) *Client {
	return &Client{
		Token:  token,
		client: sling.New().Base("https://pullapprove.com/api/"),
	}
}

// Create creates a new repository on pull approve
func (c *Client) Create(name string, org string) error {
	path := fmt.Sprintf("orgs/%s/repos/", org)
	cr := createReq{Name: name}

	req, err := c.client.Post(path).BodyJSON(&cr).Request()
	if err != nil {
		return errwrap.Wrapf("could not create the request to pull approve: {{err}}", err)
	}

	return c.doRequest(req)
}

// Delete deletes a repository from pull approve
func (c *Client) Delete(name string, org string) error {
	path := fmt.Sprintf("orgs/%s/%s/delete", org, name)
	req, err := c.client.Post(path).Request()
	if err != nil {
		return errwrap.Wrapf("could not delete the request from pull approve: {{err}}", err)
	}

	return c.doRequest(req)
}

func (c *Client) doRequest(req *http.Request) error {
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.Token))

	resp, err := c.client.Do(req, nil, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrPullApproveUnauthorized
	}

	if code := resp.StatusCode; code >= 400 {
		return ErrPullApproveServerError
	}

	return nil
}
