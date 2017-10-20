package repo

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/hellofresh/github-cli/pkg/config"
	"github.com/hellofresh/github-cli/pkg/pullapprove"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type (
	// RepoCreateor is used as an aggregator of rules to setup your repository
	RepoCreateor interface {
		Create(opts *HelloFreshRepoOpt) error
	}

	// GithubRepo contains all the hellofresh repository creation rules for github
	GithubRepo struct {
		GithubClient *github.Client
	}

	// HelloFreshRepoOpt represents the repo creation options
	HelloFreshRepoOpt struct {
		Name              string
		Org               string
		Private           bool
		PullApprove       *PullApproveRule
		Teams             *TeamsRule
		Collaborators     *CollaboratorsRule
		Labels            *LabelsRule
		Webhooks          *WebhooksRule
		BranchProtections *BranchProtectionsRule
	}

	PullApproveRule struct {
		Enabled             bool
		Client              *pullapprove.Client
		Filename            string
		ProtectedBranchName string
	}

	TeamsRule struct {
		Enabled bool
		Teams   []*config.Team
	}

	CollaboratorsRule struct {
		Enabled       bool
		Collaborators []*config.Collaborator
	}

	LabelsRule struct {
		Enabled             bool
		RemoveDefaultLabels bool
		Labels              []*config.Label
	}

	WebhooksRule struct {
		Enabled  bool
		Webhooks []*config.Webhook
	}

	BranchProtectionsRule struct {
		Enabled     bool
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

// Create aggregates all rules to create a repository with all necessary requirements
func (c *GithubRepo) Create(opts *HelloFreshRepoOpt) error {
	log.Info("Creating repository...")
	err := c.createRepo(opts.Name, opts.Org, opts.Private)
	if err != nil {
		return errors.Wrap(err, "could not create repository")
	}

	if opts.PullApprove != nil && opts.PullApprove.Enabled {
		log.Info("Adding pull approve...")
		err = c.addPullApprove(opts.Name, opts.Org, opts.PullApprove)
		if err != nil {
			return errors.Wrap(err, "could not add pull approve")
		}
	}

	if opts.Teams != nil && opts.Teams.Enabled {
		log.Info("Adding teams to repository...")
		err = c.addTeamsToRepo(opts.Name, opts.Org, opts.Teams)
		if err != nil {
			return errors.Wrap(err, "could add teams to repository")
		}
	}

	if opts.Collaborators != nil && opts.Collaborators.Enabled {
		log.Info("Adding collaborators to repository...")
		err = c.addCollaborators(opts.Name, opts.Org, opts.Collaborators)
		if err != nil {
			return errors.Wrap(err, "could not add collaborators to repository")
		}
	}

	if opts.Labels != nil && opts.Labels.Enabled {
		log.Info("Adding labels to repository...")
		err = c.addLabelsToRepo(opts.Name, opts.Org, opts.Labels)
		if err != nil {
			return errors.Wrap(err, "could add labels to repository")
		}
	}

	if opts.Webhooks != nil && opts.Webhooks.Enabled {
		log.Info("Adding webhooks to repository...")
		err = c.addWebhooksToRepo(opts.Name, opts.Org, opts.Webhooks)
		if err != nil {
			return errors.Wrap(err, "could add webhooks to repository")
		}
	}

	if opts.BranchProtections != nil && opts.BranchProtections.Enabled {
		log.Info("Adding branch protections to repository...")
		err = c.addBranchProtections(opts.Name, opts.Org, opts.BranchProtections)
		if err != nil {
			return errors.Wrap(err, "could add branch protections to repository")
		}
	}

	log.Info("Repository created!")
	return nil
}

func (c *GithubRepo) createRepo(name string, org string, private bool) error {
	repo := &github.Repository{
		Name:      github.String(name),
		Private:   github.Bool(private),
		HasIssues: github.Bool(true),
	}

	_, _, err := c.GithubClient.Repositories.Create(ctx, org, repo)
	return err
}

func (c *GithubRepo) addPullApprove(repo string, org string, opts *PullApproveRule) error {
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

func (c *GithubRepo) addTeamsToRepo(repo string, org string, opts *TeamsRule) error {
	var err error

	for _, team := range opts.Teams {
		opt := &github.OrganizationAddTeamRepoOptions{
			Permission: team.Permission,
		}

		_, err = c.GithubClient.Organizations.AddTeamRepo(ctx, team.ID, org, repo, opt)
	}

	return err
}

func (c *GithubRepo) addLabelsToRepo(repo string, org string, opts *LabelsRule) error {
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

func (c *GithubRepo) addWebhooksToRepo(repo string, org string, opts *WebhooksRule) error {
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

func (c *GithubRepo) addBranchProtections(repo string, org string, opts *BranchProtectionsRule) error {
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

func (c *GithubRepo) addCollaborators(repo string, org string, opts *CollaboratorsRule) error {
	var err error

	for _, collaborator := range opts.Collaborators {
		opt := &github.RepositoryAddCollaboratorOptions{
			Permission: collaborator.Permission,
		}

		_, err = c.GithubClient.Repositories.AddCollaborator(ctx, org, repo, collaborator.Username, opt)
	}

	return err
}
