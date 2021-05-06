package repo

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/go-github/v33/github"
	multierror "github.com/hashicorp/go-multierror"

	"github.com/hellofresh/github-cli/pkg/config"
)

type (
	// GithubRepo contains all the hellofresh repository creation Opts for github
	GithubRepo struct {
		GithubClient *github.Client
	}

	// GithubRepoOpts represents the repo creation options
	GithubRepoOpts struct {
		Teams             []*config.Team
		Collaborators     []*config.Collaborator
		Labels            *LabelsOpts
		Webhooks          []*config.Webhook
		BranchProtections config.BranchProtections
	}

	// LabelsOpts represents label options
	LabelsOpts struct {
		RemoveDefaultLabels bool
		Labels              []*config.Label
	}
)

var (
	// ErrRepositoryAlreadyExists is used when the repository already exists
	ErrRepositoryAlreadyExists = errors.New("github repository already exists")
	// ErrRepositoryLimitExceeded is used when the repository limit is exceeded
	ErrRepositoryLimitExceeded = errors.New("limit for private repos on this account is exceeded")
	// ErrLabelNotFound is used when a label is not found
	ErrLabelNotFound = errors.New("github label does not exist")
	// ErrLabeAlreadyExists is used when a label is not found
	ErrLabeAlreadyExists = errors.New("github label already exists")
	// ErrWebhookAlreadyExist is used when a webhook already exists
	ErrWebhookAlreadyExist = errors.New("github webhook already exists")
	// ErrOrganizationNotFound is used when a webhook already exists
	ErrOrganizationNotFound = errors.New("you must specify an organization to use this functionality")
)

// NewGithub creates a new instance of Client
func NewGithub(githubClient *github.Client) *GithubRepo {
	return &GithubRepo{
		GithubClient: githubClient,
	}
}

// CreateRepo creates a github repository
func (c *GithubRepo) CreateRepo(ctx context.Context, org string, repoOpts *github.Repository) (*github.Repository, error) {
	ghRepo, _, err := c.GithubClient.Repositories.Create(ctx, org, repoOpts)
	if githubError, ok := err.(*github.ErrorResponse); ok {
		if strings.Contains(githubError.Message, "Visibility can't be private") {
			err = ErrRepositoryLimitExceeded
		} else if githubError.Response.StatusCode == http.StatusUnprocessableEntity {
			err = ErrRepositoryAlreadyExists
		}
	}

	return ghRepo, err
}

// AddTeamsToRepo adds an slice of teams and their permissions to a repository
func (c *GithubRepo) AddTeamsToRepo(ctx context.Context, repo string, org string, teams []*config.Team) error {
	var err error

	if org == "" {
		return ErrOrganizationNotFound
	}

	for _, team := range teams {
		opt := &github.TeamAddTeamRepoOptions{
			Permission: team.Permission,
		}

		orgInfo, _, ghErr := c.GithubClient.Organizations.Get(ctx, org)
		if ghErr != nil {
			return err
		}

		if _, ghErr := c.GithubClient.Teams.AddTeamRepoByID(ctx, orgInfo.GetID(), int64(team.ID), "", repo, opt); ghErr != nil {
			err = multierror.Append(err, ghErr)
		}
	}

	return err
}

// AddLabelsToRepo adds an slice of labels to the repository. Optionally this can also remove github's
// default labels
func (c *GithubRepo) AddLabelsToRepo(ctx context.Context, repo string, org string, opts *LabelsOpts) error {
	var err error
	defaultLabels := []string{"bug", "duplicate", "enhancement", "help wanted", "invalid", "question", "wontfix", "good first issue"}

	for _, label := range opts.Labels {
		githubLabel := &github.Label{
			Name:  github.String(label.Name),
			Color: github.String(label.Color),
		}

		if _, _, ghErr := c.GithubClient.Issues.CreateLabel(ctx, org, repo, githubLabel); ghErr != nil {
			if internalErr, ok := ghErr.(*github.ErrorResponse); ok {
				if internalErr.Response.StatusCode == http.StatusUnprocessableEntity {
					err = multierror.Append(err, ErrLabeAlreadyExists)
				}
			}
		}
	}

	if opts.RemoveDefaultLabels {
		for _, label := range defaultLabels {
			if _, ghErr := c.GithubClient.Issues.DeleteLabel(ctx, org, repo, label); ghErr != nil {
				if internalErr, ok := ghErr.(*github.ErrorResponse); ok {
					if internalErr.Response.StatusCode == http.StatusNotFound {
						err = multierror.Append(err, ErrLabelNotFound)
					}
				}
			}
		}
	}

	return err
}

// AddWebhooksToRepo adds an slice of webhooks to the repository
func (c *GithubRepo) AddWebhooksToRepo(ctx context.Context, repo string, org string, webhooks []*config.Webhook) error {
	var err error

	for _, webhook := range webhooks {
		hook := &github.Hook{
			Config: webhook.Config,
		}
		_, _, err = c.GithubClient.Repositories.CreateHook(ctx, org, repo, hook)
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusUnprocessableEntity {
				err = multierror.Append(err, ErrWebhookAlreadyExist)
			}
		}
	}

	return err
}

// AddBranchProtections adds an slice of branch protections to the repository
func (c *GithubRepo) AddBranchProtections(ctx context.Context, repo string, org string, protections config.BranchProtections) error {
	var err error

	for branch, contexts := range protections {
		pr := &github.ProtectionRequest{
			RequiredStatusChecks: &github.RequiredStatusChecks{
				Contexts: contexts,
			},
		}
		if _, _, ghErr := c.GithubClient.Repositories.UpdateBranchProtection(ctx, org, repo, branch, pr); ghErr != nil {
			err = multierror.Append(err, ghErr)
		}
	}

	return err
}

// AddCollaborators adds an slice of collaborators and their permissions to the repository
func (c *GithubRepo) AddCollaborators(ctx context.Context, repo string, org string, collaborators []*config.Collaborator) error {
	var err error

	for _, collaborator := range collaborators {
		opt := &github.RepositoryAddCollaboratorOptions{
			Permission: collaborator.Permission,
		}

		if _, _, ghErr := c.GithubClient.Repositories.AddCollaborator(ctx, org, repo, collaborator.Username, opt); ghErr != nil {
			err = multierror.Append(err, ghErr)
		}
	}

	return err
}
