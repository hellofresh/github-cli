package cmd

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type (
	// UnseatOpts are the flags for the unseat command
	UnseatOpts struct {
		Org          string
		Page         int
		ReposPerPage int
	}
)

// NewHiringUnseat creates a new hiring unseat command
func NewHiringUnseat() *cobra.Command {
	opts := &UnseatOpts{}

	cmd := &cobra.Command{
		Use:     "unseat",
		Aliases: []string{"un"},
		Short:   "Removes external collaborators from repositories",
		Long:    `Removes external (people not in the organization) collaborators from repositories`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunUnseat(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Org, "organization", "o", "", "Github's organization")
	cmd.Flags().IntVar(&opts.ReposPerPage, "page-size", 50, "How many repositories should we get per page? (max 100)")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Starting page for repositories")

	return cmd
}

// RunUnseat runs the command to create a new hiring test repository
func RunUnseat(opts *UnseatOpts) error {
	org := opts.Org
	if org == "" {
		org = globalConfig.Github.Organization
	}
	if org == "" {
		return errors.New("Please provide an organization")
	}

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: opts.ReposPerPage, Page: opts.Page},
	}
	ctx := context.Background()
	// get all pages of results
	var allRepos []*github.Repository
	for {
		log.Infof("Fetching repositories page [%d]", opt.Page)
		repos, resp, err := githubClient.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return errors.Wrap(err, "Could not retrieve repositories")
		}

		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	for _, repo := range allRepos {
		repoName := *repo.Name

		log.Infof("Fetching collaborators for %s", repoName)
		collaborators, _, err := githubClient.Repositories.ListCollaborators(ctx, org, repoName, &github.ListCollaboratorsOptions{
			Affiliation: "outside",
		})
		if err != nil {
			return errors.Wrap(err, "Could not retrieve collaborators")
		}

		log.Infof("Deleting collaborators on %s", repoName)
		for _, collaborator := range collaborators {
			_, err := githubClient.Repositories.RemoveCollaborator(ctx, org, repoName, *collaborator.Name)
			if err != nil {
				return errors.Wrap(err, "Could not unseat collaborator")
			}
		}
	}

	log.Info("Done!")
	return nil
}
