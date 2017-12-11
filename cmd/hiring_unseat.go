package cmd

import (
	"context"
	"math"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	weekInSeconds     float64 = 604800
	weeksOfInactivity int     = 5
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
		Short:   "Removes external collaborators from repositories",
		Long:    `Removes external (people not in the organization) collaborators from repositories`,
		PreRunE: setupConnection,
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
	var unseatedCollaborators int
	ctx := context.Background()

	org := opts.Org
	if org == "" {
		org = globalConfig.GithubTestOrg.Organization
	}
	if org == "" {
		return errors.New("Please provide an organization")
	}

	allRepos, err := fetchAllRepos(org, opts.ReposPerPage, opts.Page)
	if err != nil {
		return errors.Wrap(err, "Could not retrieve repositories")
	}

	for _, repo := range allRepos {
		if isRepoInactive(repo) {
			continue
		}

		repoName := *repo.Name
		log.WithField("repo", repoName).Info("Fetching outside collaborators")
		outsideCollaborators, _, err := githubClient.Repositories.ListCollaborators(ctx, org, repoName, &github.ListCollaboratorsOptions{
			Affiliation: "outside",
		})
		if err != nil {
			return errors.Wrap(err, "Could not retrieve outside collaborators")
		}

		for _, collaborator := range outsideCollaborators {
			log.WithFields(log.Fields{
				"repo":         repoName,
				"collaborator": collaborator.GetLogin(),
			}).Info("Deleting outside collaborators")
			_, err := githubClient.Repositories.RemoveCollaborator(ctx, org, repoName, collaborator.GetLogin())
			if err != nil {
				return errors.Wrap(err, "Could not unseat outside collaborator")
			}

			unseatedCollaborators++
		}
	}

	log.Infof("Done! %d outside collaborators unseated", unseatedCollaborators)
	return nil
}

func fetchAllRepos(owner string, reposPerPage int, page int) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: reposPerPage, Page: page},
	}

	for {
		log.Infof("Fetching repositories page [%d]", opt.Page)
		repos, resp, err := githubClient.Repositories.ListByOrg(context.Background(), owner, opt)
		if err != nil {
			return allRepos, err
		}

		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func isRepoInactive(repo *github.Repository) bool {
	diff := time.Since(repo.PushedAt.Time)
	weeksAgo := roundTime(diff.Seconds() / weekInSeconds)

	return weeksAgo < weeksOfInactivity
}

func roundTime(input float64) int {
	var result float64

	if input < 0 {
		result = math.Ceil(input - 0.5)
	} else {
		result = math.Floor(input + 0.5)
	}

	// only interested in integer, ignore fractional
	i, _ := math.Modf(result)

	return int(i)
}
