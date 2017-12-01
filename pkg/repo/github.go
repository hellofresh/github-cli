package repo

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
	"github.com/hellofresh/github-cli/pkg/config"
	"github.com/hellofresh/github-cli/pkg/pullapprove"
	"github.com/pkg/errors"
)

type (
	// GithubRepo contains all the hellofresh repository creation Opts for github
	GithubRepo struct {
		GithubClient *github.Client
	}

	// GithubRepoOpts represents the repo creation options
	GithubRepoOpts struct {
		PullApprove       *PullApproveOpts
		Teams             []*config.Team
		Collaborators     []*config.Collaborator
		Labels            *LabelsOpts
		Webhooks          []*config.Webhook
		BranchProtections config.BranchProtections
	}

	// PullApproveOpts represents pull approve options
	PullApproveOpts struct {
		Client              *pullapprove.Client
		Filename            string
		ProtectedBranchName string
	}

	// LabelsOpts represents label options
	LabelsOpts struct {
		RemoveDefaultLabels bool
		Labels              []*config.Label
	}
)

var (
	ctx = context.Background()
	// ErrRepositoryAlreadyExists is used when the repository already exists
	ErrRepositoryAlreadyExists = errors.New("github repository already exists")
	// ErrRepositoryLimitExceeded is used when the repository limit is exceeded
	ErrRepositoryLimitExceeded = errors.New("limit for private repos on this account is exceeded")
	// ErrPullApproveFileAlreadyExists is used when the pull approve file already exists
	ErrPullApproveFileAlreadyExists = errors.New("github pull approve file already exists")
	// ErrLabelNotFound is used when a label is not found
	ErrLabelNotFound = errors.New("github label does not exist")
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
func (c *GithubRepo) CreateRepo(org string, repoOpts *github.Repository) (*github.Repository, error) {
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

// AddPullApprove adds a file to the github repository and calls pull approve API to register the new repo
func (c *GithubRepo) AddPullApprove(repo string, org string, opts *PullApproveOpts) error {
	var err error
	if opts.Client == nil {
		return errors.New("Cannot add pull approve, since the client is nil")
	}

	fileOpt := &github.RepositoryContentFileOptions{
		Message: github.String("Initialize repository :tada:"),
		Content: []byte("extends: hellofresh"),
		Branch:  github.String(opts.ProtectedBranchName),
	}
	_, _, err = c.GithubClient.Repositories.CreateFile(ctx, org, repo, opts.Filename, fileOpt)
	if githubError, ok := err.(*github.ErrorResponse); ok {
		if githubError.Response.StatusCode == http.StatusUnprocessableEntity {
			return ErrPullApproveFileAlreadyExists
		}
	} else {
		return err
	}

	err = opts.Client.Create(repo, org)
	if err != nil {
		return err
	}

	return err
}

// AddTeamsToRepo adds an slice of teams and their permissions to a repository
func (c *GithubRepo) AddTeamsToRepo(repo string, org string, teams []*config.Team) error {
	var err error

	if org == "" {
		return ErrOrganizationNotFound
	}

	for _, team := range teams {
		opt := &github.OrganizationAddTeamRepoOptions{
			Permission: team.Permission,
		}

		_, err = c.GithubClient.Organizations.AddTeamRepo(ctx, team.ID, org, repo, opt)
	}

	return err
}

// AddLabelsToRepo adds an slice of labels to the repository. Optionally this can also remove github's
// default labels
func (c *GithubRepo) AddLabelsToRepo(repo string, org string, opts *LabelsOpts) error {
	var err error
	defaultLabels := []string{"bug", "duplicate", "enhancement", "help wanted", "invalid", "question", "wontfix", "good first issue"}

	for _, label := range opts.Labels {
		githubLabel := &github.Label{
			Name:  github.String(label.Name),
			Color: github.String(label.Color),
		}

		_, _, err = c.GithubClient.Issues.CreateLabel(ctx, org, repo, githubLabel)
	}

	if opts.RemoveDefaultLabels {
		for _, label := range defaultLabels {
			_, err = c.GithubClient.Issues.DeleteLabel(ctx, org, repo, label)
			if githubError, ok := err.(*github.ErrorResponse); ok {
				if githubError.Response.StatusCode == http.StatusNotFound {
					err = errors.Wrap(ErrLabelNotFound, "label not found")
				}
			}
		}
	}

	return err
}

// AddWebhooksToRepo adds an slice of webhooks to the repository
func (c *GithubRepo) AddWebhooksToRepo(repo string, org string, webhooks []*config.Webhook) error {
	var err error

	for _, webhook := range webhooks {
		hook := &github.Hook{
			Name:   github.String(webhook.Type),
			Config: webhook.Config,
		}
		_, _, err = c.GithubClient.Repositories.CreateHook(ctx, org, repo, hook)
		if githubError, ok := err.(*github.ErrorResponse); ok {
			if githubError.Response.StatusCode == http.StatusUnprocessableEntity {
				err = errors.Wrap(ErrWebhookAlreadyExist, "webhook already exists")
			}
		}
	}

	return err
}

// AddBranchProtections adds an slice of branch protections to the repository
func (c *GithubRepo) AddBranchProtections(repo string, org string, protections config.BranchProtections) error {
	var err error

	for branch, contexts := range protections {
		pr := &github.ProtectionRequest{
			RequiredStatusChecks: &github.RequiredStatusChecks{
				Contexts: contexts,
			},
		}
		_, _, err = c.GithubClient.Repositories.UpdateBranchProtection(ctx, org, repo, branch, pr)
	}

	return err
}

// AddCollaborators adds an slice of collaborators and their permissions to the repository
func (c *GithubRepo) AddCollaborators(repo string, org string, collaborators []*config.Collaborator) error {
	var err error

	for _, collaborator := range collaborators {
		opt := &github.RepositoryAddCollaboratorOptions{
			Permission: collaborator.Permission,
		}

		_, err = c.GithubClient.Repositories.AddCollaborator(ctx, org, repo, collaborator.Username, opt)
	}

	return err
}
