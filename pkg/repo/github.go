package repo

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/hellofresh/github-cli/pkg/config"
	"github.com/hellofresh/github-cli/pkg/pullapprove"
	"github.com/pkg/errors"
)

type (
	// GithubRepo contains all the hellofresh repository creation Optss for github
	GithubRepo struct {
		GithubClient *github.Client
	}

	// GithubRepoOpts represents the repo creation options
	GithubRepoOpts struct {
		PullApprove       *PullApproveOpts
		Teams             *TeamsOpts
		Collaborators     *CollaboratorsOpts
		Labels            *LabelsOpts
		Webhooks          *WebhooksOpts
		BranchProtections *BranchProtectionsOpts
	}

	PullApproveOpts struct {
		Client              *pullapprove.Client
		Filename            string
		ProtectedBranchName string
	}

	TeamsOpts struct {
		Teams []*config.Team
	}

	CollaboratorsOpts struct {
		Collaborators []*config.Collaborator
	}

	LabelsOpts struct {
		RemoveDefaultLabels bool
		Labels              []*config.Label
	}

	WebhooksOpts struct {
		Webhooks []*config.Webhook
	}

	BranchProtectionsOpts struct {
		Protections map[string][]string
	}
)

var ctx = context.Background()

// NewGithub creates a new instance of Client
func NewGithub(githubClient *github.Client) *GithubRepo {
	return &GithubRepo{
		GithubClient: githubClient,
	}
}

func (c *GithubRepo) CreateRepo(name string, description string, org string, private bool) error {
	repo := &github.Repository{
		Name:        github.String(name),
		Description: github.String(description),
		Private:     github.Bool(private),
		HasIssues:   github.Bool(true),
	}

	_, _, err := c.GithubClient.Repositories.Create(ctx, org, repo)
	return err
}

func (c *GithubRepo) AddPullApprove(repo string, org string, opts *PullApproveOpts) error {
	if opts.Client == nil {
		return errors.New("Cannot add pull approve, since the client is nil")
	}

	fileOpt := &github.RepositoryContentFileOptions{
		Message: github.String("Initialize repository :tada:"),
		Content: []byte("extends: hellofresh"),
		Branch:  github.String(opts.ProtectedBranchName),
	}
	_, _, err := c.GithubClient.Repositories.CreateFile(ctx, org, repo, opts.Filename, fileOpt)
	if err != nil {
		return err
	}

	err = opts.Client.Create(repo, org)
	if err != nil {
		return err
	}

	return nil
}

func (c *GithubRepo) AddTeamsToRepo(repo string, org string, opts *TeamsOpts) error {
	var err error

	for _, team := range opts.Teams {
		opt := &github.OrganizationAddTeamRepoOptions{
			Permission: team.Permission,
		}

		_, err = c.GithubClient.Organizations.AddTeamRepo(ctx, team.ID, org, repo, opt)
	}

	return err
}

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
		}
	}

	return err
}

func (c *GithubRepo) AddWebhooksToRepo(repo string, org string, opts *WebhooksOpts) error {
	var err error

	for _, webhook := range opts.Webhooks {
		hook := &github.Hook{
			Name:   github.String(webhook.Type),
			Config: webhook.Config,
		}
		_, _, err = c.GithubClient.Repositories.CreateHook(ctx, org, repo, hook)
	}

	return err
}

func (c *GithubRepo) AddBranchProtections(repo string, org string, opts *BranchProtectionsOpts) error {
	var err error

	for branch, contexts := range opts.Protections {
		pr := &github.ProtectionRequest{
			RequiredStatusChecks: &github.RequiredStatusChecks{
				Contexts: contexts,
			},
		}
		_, _, err = c.GithubClient.Repositories.UpdateBranchProtection(ctx, org, repo, branch, pr)
	}

	return err
}

func (c *GithubRepo) AddCollaborators(repo string, org string, opts *CollaboratorsOpts) error {
	var err error

	for _, collaborator := range opts.Collaborators {
		opt := &github.RepositoryAddCollaboratorOptions{
			Permission: collaborator.Permission,
		}

		_, err = c.GithubClient.Repositories.AddCollaborator(ctx, org, repo, collaborator.Username, opt)
	}

	return err
}
