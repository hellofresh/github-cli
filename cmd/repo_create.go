package cmd

import (
	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/hellofresh/github-cli/pkg/pullapprove"
	"github.com/hellofresh/github-cli/pkg/repo"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// CreateRepoOptions are the flags for the create repository command
type CreateRepoOptions struct {
	Description          string
	Org                  string
	Private              bool
	HasPullApprove       bool
	HasTeams             bool
	HasCollaborators     bool
	HasLabels            bool
	HasDefaultLabels     bool
	HasWebhooks          bool
	HasBranchProtections bool
	HasIssues            bool
	HasWiki              bool
	HasPages             bool
}

// NewCreateRepoCmd creates a new create repo command
func NewCreateRepoCmd() *cobra.Command {
	opts := &CreateRepoOptions{}

	cmd := &cobra.Command{
		Use:     "create [name]",
		Aliases: []string{"cr"},
		Short:   "Creates a new github repository",
		Long:    `Creates a new github repository based on the rules defined on your .github.toml`,
		PreRunE: setupConnection,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunCreateRepo(args[0], opts)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 || args[0] == "" {
				return errors.New("Please provide a repository name")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "The repository's description")
	cmd.Flags().StringVarP(&opts.Org, "organization", "o", "", "Github's organization")
	cmd.Flags().BoolVar(&opts.Private, "private", true, "Is the repository private?")

	cmd.Flags().BoolVar(&opts.HasIssues, "has-issues", true, "Enables issue pages")
	cmd.Flags().BoolVar(&opts.HasWiki, "has-wiki", false, "Enables wiki pages?")
	cmd.Flags().BoolVar(&opts.HasPages, "has-pages", false, "Enables github pages?")
	cmd.Flags().BoolVar(&opts.HasPullApprove, "has-pullapprove", true, "Enables pull approve")
	cmd.Flags().BoolVar(&opts.HasTeams, "has-teams", true, "Enable teams")
	cmd.Flags().BoolVar(&opts.HasLabels, "has-labels", true, "Enable labels")
	cmd.Flags().BoolVar(&opts.HasDefaultLabels, "rm-default-labels", true, "Removes the default github labels")
	cmd.Flags().BoolVar(&opts.HasWebhooks, "has-webhooks", false, "Enables webhooks configurations")
	cmd.Flags().BoolVar(&opts.HasBranchProtections, "has-branch-protections", true, "Enables branch protections")

	return cmd
}

// RunCreateRepo runs the command to create a new repository
func RunCreateRepo(repoName string, opts *CreateRepoOptions) error {
	var err error
	var org string

	description := opts.Description

	org = opts.Org
	if org == "" {
		org = globalConfig.Github.Organization
	}

	githubOpts := &repo.GithubRepoOpts{
		PullApprove: &repo.PullApproveOpts{
			Filename:            globalConfig.PullApprove.Filename,
			ProtectedBranchName: globalConfig.PullApprove.ProtectedBranchName,
			Client:              pullapprove.New(globalConfig.PullApprove.Token),
		},
		Labels: &repo.LabelsOpts{
			RemoveDefaultLabels: globalConfig.Github.RemoveDefaultLabels,
			Labels:              globalConfig.Github.Labels,
		},
		Teams:             globalConfig.Github.Teams,
		Webhooks:          globalConfig.Github.Webhooks,
		BranchProtections: globalConfig.Github.Protections,
	}

	creator := repo.NewGithub(githubClient)

	log.Info("Creating repository...")
	ghRepo, err := creator.CreateRepo(org, &github.Repository{
		Name:        github.String(repoName),
		Description: github.String(description),
		Private:     github.Bool(opts.Private),
		HasIssues:   github.Bool(opts.HasIssues),
		HasWiki:     github.Bool(opts.HasWiki),
		HasPages:    github.Bool(opts.HasPages),
		AutoInit:    github.Bool(true),
	})
	if errors.Cause(err) == repo.ErrRepositoryAlreadyExists {
		log.Info("Repository already exists. Trying to normalize it...")
	} else if err != nil {
		return errors.Wrap(err, "Could not create repository")
	}

	if opts.HasPullApprove {
		log.Info("Adding pull approve...")
		err = creator.AddPullApprove(repoName, org, githubOpts.PullApprove)
		if errors.Cause(err) == repo.ErrPullApproveFileAlreadyExists {
			color.Cyan("Pull approve already exists, moving on...")
		} else if err != nil {
			return errors.Wrap(err, "Could not add pull approve")
		}
	}

	if opts.HasTeams {
		log.Info("Adding teams to repository...")
		err = creator.AddTeamsToRepo(repoName, org, githubOpts.Teams)
		if err != nil {
			return errors.Wrap(err, "Could not add teams to repository")
		}
	}

	if opts.HasCollaborators {
		log.Info("Adding collaborators to repository...")
		err = creator.AddCollaborators(repoName, org, githubOpts.Collaborators)
		if err != nil {
			return errors.Wrap(err, "Could not add collaborators to repository")
		}
	}

	if opts.HasLabels {
		log.Info("Adding labels to repository...")
		err = creator.AddLabelsToRepo(repoName, org, githubOpts.Labels)
		if errors.Cause(err) == repo.ErrLabelNotFound {
			color.Cyan("Default labels does not exists, moving on...")
		} else if err != nil {
			return errors.Wrap(err, "Could not add labels to repository")
		}
	}

	if opts.HasWebhooks {
		log.Info("Adding webhooks to repository...")
		err = creator.AddWebhooksToRepo(repoName, org, githubOpts.Webhooks)
		if errors.Cause(err) == repo.ErrWebhookAlreadyExist {
			color.Cyan("Webhook already exists, moving on...")
		} else if err != nil {
			return errors.Wrap(err, "Could not add webhooks to repository")
		}
	}

	if opts.HasBranchProtections {
		log.Info("Adding branch protections to repository...")
		err = creator.AddBranchProtections(repoName, org, githubOpts.BranchProtections)
		if err != nil {
			return errors.Wrap(err, "Could not add branch protections to repository")
		}
	}

	if ghRepo != nil {
		log.Infof("Repository created! \n Here is how to access it %s", ghRepo.GetGitURL())
	} else {
		log.Info("Repository normalized!")
	}

	return nil
}
