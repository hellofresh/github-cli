package cmd

import (
	"context"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

type (
	// UnseatFlags are the flags for the unseat command
	UnseatFlags struct {
		Org          string
		Page         int
		ReposPerPage int
	}
)

var unseatFlags UnseatFlags

func init() {
	unseatCmd.Flags().StringVarP(&unseatFlags.Org, "organization", "o", "", "Github's organization")
	unseatCmd.Flags().IntVar(&unseatFlags.ReposPerPage, "page-size", 50, "How many repositories should we get per page? (max 100)")
	unseatCmd.Flags().IntVar(&unseatFlags.Page, "page", 1, "Starting page for repositories")
}

// RunUnseat runs the command to create a new hiring test repository
func RunUnseat(cmd *cobra.Command, args []string) {
	org := unseatFlags.Org
	if org == "" {
		org = globalConfig.Github.Organization
	}
	checkEmpty(org, "Please provide an organization")

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: unseatFlags.ReposPerPage, Page: unseatFlags.Page},
	}
	ctx := context.Background()
	// get all pages of results
	var allRepos []*github.Repository
	for {
		color.White("Fetching repositories page [%d]", opt.Page)
		repos, resp, err := githubClient.Repositories.ListByOrg(ctx, org, opt)
		checkEmpty(err, "Could not retrieve repositories")

		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	for _, repo := range allRepos {
		repoName := *repo.Name

		color.White("Fetching collaborators for %s", repoName)
		collaborators, _, err := githubClient.Repositories.ListCollaborators(ctx, org, repoName, &github.ListCollaboratorsOptions{
			Affiliation: "outside",
		})
		checkEmpty(err, "Could not retrieve collaborators")

		color.White("Deleting collaborators on %s", repoName)
		for _, collaborator := range collaborators {
			_, err := githubClient.Repositories.RemoveCollaborator(ctx, org, repoName, *collaborator.Name)
			checkEmpty(err, "Could not unseat collaborator")
		}
	}

	color.Green("Done!")
}
