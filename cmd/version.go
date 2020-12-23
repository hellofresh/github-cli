package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/hellofresh/github-cli/pkg/log"
)

var version = "0.0.0-dev"

// NewVersionCmd creates a new version command
func NewVersionCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Print the version information",
		Aliases: []string{"v"},
		Run: func(cmd *cobra.Command, args []string) {
			logger := log.WithContext(ctx)
			logger.Infof("github-cli %s", version)
		},
	}
}
