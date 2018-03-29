package cmd

import (
	"context"
	"fmt"

	"github.com/hashicorp/errwrap"
	"github.com/hellofresh/github-cli/pkg/log"
	"github.com/italolelis/goupdater"
	"github.com/spf13/cobra"
)

const (
	githubOwner = "hellofresh"
	githubRepo  = "github-cli"
)

// UpdateOptions are the command flags
type UpdateOptions struct{}

// NewUpdateCmd creates a new update command
func NewUpdateCmd(ctx context.Context) *cobra.Command {
	opts := &UpdateOptions{}

	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"self-update"},
		Short:   fmt.Sprintf("Check for new versions of %s", githubRepo),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunUpdate(ctx, opts)
		},
	}

	return cmd
}

// RunUpdate runs the update command
func RunUpdate(ctx context.Context, opts *UpdateOptions) error {
	logger := log.WithContext(ctx)
	logger.Info("Checking if any new version is available...")

	resolver, err := goupdater.NewGithubWithContext(ctx, goupdater.GithubOpts{
		Owner: githubOwner,
		Repo:  githubRepo,
	})
	if err != nil {
		return errwrap.Wrapf("could not create the updater client: {{err}}", err)
	}

	updated, err := goupdater.UpdateWithContext(ctx, resolver, version)
	if err != nil {
		return errwrap.Wrapf("could not update the binary: {{err}}", err)
	}

	if updated {
		logger.Infof("You are now using the latest version of %s", githubRepo)
	} else {
		logger.Infof("You already have the latest version of %s", githubRepo)
	}

	return nil
}
