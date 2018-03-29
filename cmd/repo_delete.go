package cmd

import (
	"context"
	"errors"

	"github.com/hashicorp/errwrap"
	"github.com/hellofresh/github-cli/pkg/config"
	gh "github.com/hellofresh/github-cli/pkg/github"
	"github.com/hellofresh/github-cli/pkg/log"
	"github.com/spf13/cobra"
)

type (
	// DeleteRepoOpts are the flags for the delete repo command
	DeleteRepoOpts struct {
		Name string
	}
)

// NewDeleteRepoCmd creates a new delete repo command
func NewDeleteRepoCmd(ctx context.Context) *cobra.Command {
	opts := &DeleteRepoOpts{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Deletes a github repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunDeleteRepo(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "The name of the repository")

	return cmd
}

// RunDeleteRepo runs the command to delete a repository
func RunDeleteRepo(ctx context.Context, opts *DeleteRepoOpts) error {
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

	_, err := githubClient.Repositories.Delete(ctx, org, opts.Name)
	if err != nil {
		return errwrap.Wrapf("Could not delete repository: {{err}}", err)
	}

	logger.Info("Repository %s deleted!", opts.Name)

	return nil
}
