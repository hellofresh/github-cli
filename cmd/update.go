package cmd

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/hellofresh/updater-go/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/hellofresh/github-cli/pkg/log"
)

const (
	githubOwner = "hellofresh"
	githubRepo  = "github-cli"
)

// UpdateOptions are the command flags
type UpdateOptions struct {
	timeout time.Duration
}

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

	cmd.PersistentFlags().DurationVar(&opts.timeout, "timeout", 10*time.Second, "Request timeout when searching for new release")

	return cmd
}

// RunUpdate runs the update command
func RunUpdate(ctx context.Context, opts *UpdateOptions) error {
	logger := log.WithContext(ctx)
	logger.Info("Checking if any new version is available...")

	resolver := updater.NewGithubClient(
		githubOwner,
		githubRepo,
		"",
		updater.StableRelease,
		func(asset string) bool {
			matchesFilter := strings.Contains(asset, fmt.Sprintf("_%s_%s", runtime.GOOS, runtime.GOARCH))

			logger.WithFields(logrus.Fields{
				"asset":    asset,
				"filtered": matchesFilter,
			}).Debug("Filtering release asset")

			return matchesFilter
		},
		opts.timeout,
	)

	updateTo, err := updater.LatestRelease(resolver)
	if rootErr := errors.Unwrap(err); rootErr == updater.ErrNoRepository {
		return fmt.Errorf("unable to acceess %s/%s repository", githubOwner, githubRepo)
	}
	if err != nil {
		return fmt.Errorf("could not retrieve release for update: %w", err)
	}

	if updateTo.Name == version {
		logger.Infof("You already have the latest version of %s/%s", githubOwner, githubRepo)
		return nil
	}

	if err := updater.SelfUpdate(updateTo); err != nil {
		return fmt.Errorf("could not update release to version %q", updateTo.Name)
	}

	logger.Infof("Updated to the version %s", updateTo.Name)

	return nil
}
