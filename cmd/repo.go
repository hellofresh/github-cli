package cmd

import "github.com/spf13/cobra"

// NewRepoCmd aggregates the repo comamnds
func NewRepoCmd() *cobra.Command {
	// Repo commands
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Github repository management",
	}

	cmd.AddCommand(NewCreateRepoCmd())
	cmd.AddCommand(NewDeleteRepoCmd())

	return cmd
}
