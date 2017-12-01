package cmd

import (
	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/hellofresh/github-cli/pkg/pullapprove"
	"github.com/hellofresh/github-cli/pkg/repo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type (
	// CreateRepoFlags are the flags for the create repository command
	CreateRepoFlags struct {
		Name                 string
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
)

var createRepoFlags CreateRepoFlags

func init() {
	repoCmd.AddCommand(createRepoCmd)

	createRepoCmd.Flags().StringVarP(&createRepoFlags.Name, "name", "n", "", "The name of the repository")
	createRepoCmd.Flags().StringVarP(&createRepoFlags.Description, "description", "d", "", "The repository's description")
	createRepoCmd.Flags().StringVarP(&createRepoFlags.Org, "organization", "o", "", "Github's organization")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.Private, "private", true, "Is the repository private?")

	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasIssues, "has-issues", true, "Enables issue pages")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasWiki, "has-wiki", false, "Enables wiki pages?")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasPages, "has-pages", false, "Enables github pages?")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasPullApprove, "has-pullapprove", true, "Enables pull approve")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasTeams, "has-teams", true, "Enable teams")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasLabels, "has-labels", true, "Enable labels")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasDefaultLabels, "rm-default-labels", true, "Removes the default github labels")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasWebhooks, "has-webhooks", false, "Enables webhooks configurations")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasBranchProtections, "has-branch-protections", true, "Enables branch protections")
}

// RunCreateRepo runs the command to create a new repository
func RunCreateRepo(cmd *cobra.Command, args []string) {
	var err error
	var org string

	name := createRepoFlags.Name
	checkEmpty(name, "Please provide a repository name")

	description := createRepoFlags.Description

	org = createRepoFlags.Org
	if org == "" {
		org = globalConfig.Github.Organization
	}

	opts := &repo.GithubRepoOpts{
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

	color.White("Creating repository...")
	ghRepo, err := creator.CreateRepo(org, &github.Repository{
		Name:        github.String(name),
		Description: github.String(description),
		Private:     github.Bool(createRepoFlags.Private),
		HasIssues:   github.Bool(createRepoFlags.HasIssues),
		HasWiki:     github.Bool(createRepoFlags.HasWiki),
		HasPages:    github.Bool(createRepoFlags.HasPages),
		AutoInit:    github.Bool(true),
	})
	if errors.Cause(err) == repo.ErrRepositoryAlreadyExists {
		color.Cyan("Repository already exists. Trying to normalize it...")
	} else {
		checkEmpty(errors.Wrap(err, "could not create repository"), "")
	}

	if createRepoFlags.HasPullApprove {
		color.White("Adding pull approve...")
		err = creator.AddPullApprove(name, org, opts.PullApprove)
		if errors.Cause(err) == repo.ErrPullApproveFileAlreadyExists {
			color.Cyan("Pull approve already exists, moving on...")
		} else {
			checkEmpty(errors.Wrap(err, "could not add pull approve"), "")
		}
	}

	if createRepoFlags.HasTeams {
		color.White("Adding teams to repository...")
		err = creator.AddTeamsToRepo(name, org, opts.Teams)
		checkEmpty(errors.Wrap(err, "could not add teams to repository"), "")
	}

	if createRepoFlags.HasCollaborators {
		color.White("Adding collaborators to repository...")
		err = creator.AddCollaborators(name, org, opts.Collaborators)
		checkEmpty(errors.Wrap(err, "could not add collaborators to repository"), "")
	}

	if createRepoFlags.HasLabels {
		color.White("Adding labels to repository...")
		err = creator.AddLabelsToRepo(name, org, opts.Labels)
		if errors.Cause(err) == repo.ErrLabelNotFound {
			color.Cyan("Default labels does not exists, moving on...")
		} else {
			checkEmpty(errors.Wrap(err, "could not add labels to repository"), "")
		}
	}

	if createRepoFlags.HasWebhooks {
		color.White("Adding webhooks to repository...")
		err = creator.AddWebhooksToRepo(name, org, opts.Webhooks)
		if errors.Cause(err) == repo.ErrWebhookAlreadyExist {
			color.Cyan("Webhook already exists, moving on...")
		} else {
			checkEmpty(errors.Wrap(err, "could not add webhooks to repository"), "")
		}
	}

	if createRepoFlags.HasBranchProtections {
		color.White("Adding branch protections to repository...")
		err = creator.AddBranchProtections(name, org, opts.BranchProtections)
		checkEmpty(errors.Wrap(err, "could not add branch protections to repository"), "")
	}

	if ghRepo != nil {
		color.Green("Repository created! \n Here is how to access it %s", *ghRepo.URL)
	} else {
		color.Green("Repository normalized!")
	}
}
