package pullapprove

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// Client represents a pull approve client
type Client struct {
	BaseURL string
	Token   string
}

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
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/orgs/%s/repos/", c.BaseURL, organization), bytes.NewBuffer(body))

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.Token))
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil && res.StatusCode != http.StatusOK {
		return errors.Wrap(err, "Could not create a pull approve repository")
	}

	return nil
}
