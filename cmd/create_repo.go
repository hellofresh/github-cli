package cmd

import (
	"github.com/fatih/color"
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
	}
)

var createRepoFlags CreateRepoFlags

func init() {
	repoCmd.AddCommand(createRepoCmd)

	createRepoCmd.Flags().StringVarP(&createRepoFlags.Name, "name", "n", "", "The name of the repository")
	createRepoCmd.Flags().StringVarP(&createRepoFlags.Description, "description", "d", "", "The repository's description")
	createRepoCmd.Flags().StringVarP(&createRepoFlags.Org, "organization", "o", "", "Github's organization")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.Private, "private", true, "Is the repository private?")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasPullApprove, "add-pullapprove", true, "Enables pull approve")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasTeams, "add-teams", true, "Enable teams")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasLabels, "add-labels", true, "Enable labels")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasDefaultLabels, "add-default-labels", true, "Removes the default github labels")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasWebhooks, "add-webhooks", true, "Enables webhooks configurations")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasBranchProtections, "add-branch-protections", true, "Enables branch protections")
}

// RunCreateRepo runs the command to create a new repository
func RunCreateRepo(cmd *cobra.Command, args []string) {
	var err error

	name := createRepoFlags.Name
	checkEmpty(name, "Please provide a repository name")

	description := createRepoFlags.Description

	org := createRepoFlags.Org
	if org == "" {
		org = globalConfig.Github.Organization
	}
	checkEmpty(org, "Please provide an organization")

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
		Teams: &repo.TeamsOpts{
			Teams: globalConfig.Github.Teams,
		},
		Webhooks: &repo.WebhooksOpts{
			Webhooks: globalConfig.Github.Webhooks,
		},
		BranchProtections: &repo.BranchProtectionsOpts{
			Protections: globalConfig.Github.Protections,
		},
	}

	creator := repo.NewGithub(githubClient)

	color.White("Creating repository...")
	err = creator.CreateRepo(name, description, org, createRepoFlags.Private)
	checkEmpty(errors.Wrap(err, "could not create repository"), "")

	if createRepoFlags.HasPullApprove {
		color.White("Adding pull approve...")
		err = creator.AddPullApprove(name, org, opts.PullApprove)
		checkEmpty(errors.Wrap(err, "could not add pull approve"), "")
	}

	if createRepoFlags.HasTeams {
		color.White("Adding teams to repository...")
		err = creator.AddTeamsToRepo(name, org, opts.Teams)
		checkEmpty(errors.Wrap(err, "could add teams to repository"), "")
	}

	if createRepoFlags.HasCollaborators {
		color.White("Adding collaborators to repository...")
		err = creator.AddCollaborators(name, org, opts.Collaborators)
		checkEmpty(errors.Wrap(err, "could not add collaborators to repository"), "")
	}

	if createRepoFlags.HasLabels {
		color.White("Adding labels to repository...")
		err = creator.AddLabelsToRepo(name, org, opts.Labels)
		checkEmpty(errors.Wrap(err, "could add labels to repository"), "")
	}

	if createRepoFlags.HasWebhooks {
		color.White("Adding webhooks to repository...")
		err = creator.AddWebhooksToRepo(name, org, opts.Webhooks)
		checkEmpty(errors.Wrap(err, "could add webhooks to repository"), "")
	}

	if createRepoFlags.HasBranchProtections {
		color.White("Adding branch protections to repository...")
		err = creator.AddBranchProtections(name, org, opts.BranchProtections)
		checkEmpty(errors.Wrap(err, "could add branch protections to repository"), "")
	}

	color.Green("Repository created!")
	checkEmpty(err, "Could not create github repo")
}