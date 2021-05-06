package cmd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/hellofresh/github-cli/pkg/config"
	gh "github.com/hellofresh/github-cli/pkg/github"
	"github.com/hellofresh/github-cli/pkg/log"
)

const (
	weekInSeconds     float64 = 604800
	weeksOfInactivity int     = 5
)

type (
	// UnseatOpts are the flags for the unseat command
	UnseatOpts struct {
		Page         int
		ReposPerPage int
	}
)

// NewHiringUnseat creates a new hiring unseat command
func NewHiringUnseat(ctx context.Context) *cobra.Command {
	opts := &UnseatOpts{}

	cmd := &cobra.Command{
		Use:   "unseat",
		Short: "Removes external collaborators from repositories",
		Long:  `Removes external (people not in the organization) collaborators from repositories`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunUnseat(ctx, opts)
		},
	}

	cmd.Flags().IntVar(&opts.ReposPerPage, "page-size", 50, "How many repositories should we get per page? (max 100)")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Starting page for repositories")

	return cmd
}

// RunUnseat runs the command to create a new hiring test repository
func RunUnseat(ctx context.Context, opts *UnseatOpts) error {
	var unseatedCollaborators int

	logger := log.WithContext(ctx)
	cfg := config.WithContext(ctx)
	githubClient := gh.WithContext(ctx)
	if githubClient == nil {
		return errors.New("failed to get github client")
	}

	org := cfg.GithubTestOrg.Organization
	if org == "" {
		return errors.New("please provide an organization")
	}

	logger.Info("Fetching repositories...")
	allRepos, err := fetchAllRepos(ctx, org, opts.ReposPerPage, opts.Page)
	if err != nil {
		return fmt.Errorf("could not retrieve repositories: %w", err)
	}
	logger.Infof("%d repositories fetched!", len(allRepos))

	logger.Info("Removing outside colaborators...")
	for _, repo := range allRepos {
		if isRepoInactive(repo) {
			continue
		}

		repoName := *repo.Name
		logger.WithField("repo", repoName).Debug("Fetching outside collaborators")
		outsideCollaborators, _, err := githubClient.Repositories.ListCollaborators(ctx, org, repoName, &github.ListCollaboratorsOptions{
			Affiliation: "outside",
		})
		if err != nil {
			return fmt.Errorf("could not retrieve outside collaborators: %w", err)
		}

		for _, collaborator := range outsideCollaborators {
			logger.WithFields(logrus.Fields{
				"repo":         repoName,
				"collaborator": collaborator.GetLogin(),
			}).Info("Deleting outside collaborators")
			_, err := githubClient.Repositories.RemoveCollaborator(ctx, org, repoName, collaborator.GetLogin())
			if err != nil {
				return fmt.Errorf("could not unseat outside collaborator: %w", err)
			}

			unseatedCollaborators++
		}
	}

	logger.Infof("Done! %d outside collaborators unseated", unseatedCollaborators)
	return nil
}

func fetchAllRepos(ctx context.Context, owner string, reposPerPage int, page int) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	logger := log.WithContext(ctx)
	githubClient := gh.WithContext(ctx)
	if githubClient == nil {
		return nil, errors.New("failed to get github client")
	}

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: reposPerPage, Page: page},
	}

	for {
		logger.Debugf("Fetching repositories page [%d]", opt.Page)
		repos, resp, err := githubClient.Repositories.ListByOrg(ctx, owner, opt)
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
