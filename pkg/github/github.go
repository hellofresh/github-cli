package github

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type githubKeyType int

const githubKey githubKeyType = iota

// NewContext returns a context with the github client imported
func NewContext(ctx context.Context, token string) (context.Context, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return context.WithValue(ctx, githubKey, github.NewClient(tc)), nil
}

// WithContext returns a github client from the context
func WithContext(ctx context.Context) *github.Client {
	if ctx == nil {
		return nil
	}

	if ctxKube, ok := ctx.Value(githubKey).(*github.Client); ok {
		return ctxKube
	}

	return nil
}
