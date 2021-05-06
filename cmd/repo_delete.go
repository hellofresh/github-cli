package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hellofresh/github-cli/pkg/config"
	gh "github.com/hellofresh/github-cli/pkg/github"
	"github.com/hellofresh/github-cli/pkg/log"
)

type (
	// DeleteRepoOpts are the flags for the delete repo command
	DeleteRepoOpts struct{}
)

// NewDeleteRepoCmd creates a new delete repo command
func NewDeleteRepoCmd(ctx context.Context) *cobra.Command {
	opts := &DeleteRepoOpts{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Deletes a github repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunDeleteRepo(ctx, args[0], opts)
		},
		Args: cobra.MinimumNArgs(1),
	}

	return cmd
}

// RunDeleteRepo runs the command to delete a repository
func RunDeleteRepo(ctx context.Context, name string, opts *DeleteRepoOpts) error {
	logger := log.WithContext(ctx)
	cfg := config.WithContext(ctx)
	githubClient := gh.WithContext(ctx)
	if githubClient == nil {
		return errors.New("failed to get github client")
	}

	org := cfg.Github.Organization
	if org == "" {
		return errors.New("please provide an organization")
	}

	logger.Debug("Fetching repo details from Github")
	_, _, err := githubClient.Repositories.Get(ctx, org, name)

	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return fmt.Errorf("github repo does not exist or you do not have access: %w", err)
		}

		return fmt.Errorf("unexpected error when tried to get a repository info: %w", err)
	}

	_, err = githubClient.Repositories.Delete(ctx, org, name)
	if err != nil {
		return fmt.Errorf("could not delete repository: %w", err)
	}

	logger.Infof("Repository %s deleted!", name)

	return nil
}
