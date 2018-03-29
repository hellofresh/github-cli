package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

// NewRepoCmd aggregates the repo comamnds
func NewRepoCmd(ctx context.Context) *cobra.Command {
	// Repo commands
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Github repository management",
	}

	cmd.AddCommand(NewCreateRepoCmd(ctx))
	cmd.AddCommand(NewDeleteRepoCmd(ctx))

	return cmd
}
