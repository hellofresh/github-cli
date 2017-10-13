package github

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/hellofresh/github-cli/pkg/config"
	"github.com/hellofresh/github-cli/pkg/pullapprove"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type (
	// RepoCreateor is used as an aggregator of rules to setup your repository
	RepoCreateor interface {
		Create() error
	}

	// HelloFreshRepoCreator contains all the hellofresh repository creation rules
	HelloFreshRepoCreator struct {
		GithubClient *github.Client
		Name         string
		Org          string
		Private      bool
		opts         *HelloFreshRepoOpt
	}

	// HelloFreshRepoOpt represents the repo creation options
	HelloFreshRepoOpt struct {
		Token             string
		Private           bool
		PullApprove       *PullApproveRule
		Teams             *TeamsRule
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

// NewHelloFreshRepoCreator creates a new instance of Client
func NewHelloFreshRepoCreator(name string, org string, opts *HelloFreshRepoOpt) (*HelloFreshRepoCreator, error) {
	if opts.Token == "" {
		return nil, errors.New("You must provide a github token")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: opts.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &HelloFreshRepoCreator{
		Name:         name,
		Org:          org,
		Private:      opts.Private,
		GithubClient: github.NewClient(tc),
		opts:         opts,
	}, nil
}

// Create aggregates all rules to create a repository with all necessary requirements
func (c *HelloFreshRepoCreator) Create() error {
	log.Info("Creating repository...")
	err := c.createRepo()
	if err != nil {
		return errors.Wrap(err, "could not create repository")
	}

	if c.opts.PullApprove.Enabled {
		log.Info("Adding pull approve...")
		err = c.addPullApprove()
		if err != nil {
			return errors.Wrap(err, "could not add pull approve")
		}
	}

	if c.opts.Teams.Enabled {
		log.Info("Adding teams to repository...")
		err = c.addTeamsToRepo()
		if err != nil {
			return errors.Wrap(err, "could add teams to repository")
		}
	}

	if c.opts.Labels.Enabled {
		log.Info("Adding labels to repository...")
		err = c.addLabelsToRepo()
		if err != nil {
			return errors.Wrap(err, "could add labels to repository")
		}
	}

	if c.opts.Webhooks.Enabled {
		log.Info("Adding webhooks to repository...")
		err = c.addWebhooksToRepo()
		if err != nil {
			return errors.Wrap(err, "could add webhooks to repository")
		}
	}

	if c.opts.BranchProtections.Enabled {
		log.Info("Adding branch protections to repository...")
		err = c.addBranchProtections()
		if err != nil {
			return errors.Wrap(err, "could add branch protections to repository")
		}
	}

	log.Info("Repository created!")
	return nil
}

func (c *HelloFreshRepoCreator) createRepo() error {
	repo := &github.Repository{
		Name:      github.String(c.Name),
		Private:   github.Bool(c.Private),
		HasIssues: github.Bool(true),
	}

	_, _, err := c.GithubClient.Repositories.Create(ctx, c.Org, repo)
	return err
}

func (c *HelloFreshRepoCreator) addPullApprove() error {
	if c.opts.PullApprove.Client == nil {
		return errors.New("Cannot add pull approve, since the client is nil")
	}

	fileOpt := &github.RepositoryContentFileOptions{
		Message: github.String("Initialize repository :tada:"),
		Content: []byte("extends: hellofresh"),
		Branch:  github.String(c.opts.PullApprove.ProtectedBranchName),
	}
	_, _, err := c.GithubClient.Repositories.CreateFile(ctx, c.Org, c.Name, c.opts.PullApprove.Filename, fileOpt)
	if err != nil {
		return err
	}

	err = c.opts.PullApprove.Client.Create(c.Name, c.Org)
	if err != nil {
		return err
	}

	return nil
}

func (c *HelloFreshRepoCreator) addTeamsToRepo() error {
	var err error

	for _, team := range c.opts.Teams.Teams {
		opt := &github.OrganizationAddTeamRepoOptions{
			Permission: team.Permission,
		}

		_, err = c.GithubClient.Organizations.AddTeamRepo(ctx, team.ID, c.Org, c.Name, opt)
	}

	return err
}

func (c *HelloFreshRepoCreator) addLabelsToRepo() error {
	var err error
	defaultLabels := []string{"bug", "duplicate", "enhancement", "help wanted", "invalid", "question", "wontfix", "good first issue"}

	for _, label := range c.opts.Labels.Labels {
		githubLabel := &github.Label{
			Name:  github.String(label.Name),
			Color: github.String(label.Color),
		}

		_, _, err = c.GithubClient.Issues.CreateLabel(ctx, c.Org, c.Name, githubLabel)
	}

	if c.opts.Labels.RemoveDefaultLabels {
		for _, label := range defaultLabels {
			_, err = c.GithubClient.Issues.DeleteLabel(ctx, c.Org, c.Name, label)
		}
	}

	return err
}

func (c *HelloFreshRepoCreator) addWebhooksToRepo() error {
	var err error

	for _, webhook := range c.opts.Webhooks.Webhooks {
		hook := &github.Hook{
			Name:   github.String(webhook.Type),
			Config: webhook.Config,
		}
		_, _, err = c.GithubClient.Repositories.CreateHook(ctx, c.Org, c.Name, hook)
	}

	return err
}

func (c *HelloFreshRepoCreator) addBranchProtections() error {
	var err error

	for branch, contexts := range c.opts.BranchProtections.Protections {
		pr := &github.ProtectionRequest{
			RequiredStatusChecks: &github.RequiredStatusChecks{
				Contexts: contexts,
			},
		}
		_, _, err = c.GithubClient.Repositories.UpdateBranchProtection(ctx, c.Org, c.Name, branch, pr)
	}

	return err
}
