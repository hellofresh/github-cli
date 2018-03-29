package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

// NewHiringCmd aggregates the hiring comamnds
func NewHiringCmd(ctx context.Context) *cobra.Command {
	// Repo commands
	cmd := &cobra.Command{
		Use:   "hiring",
		Short: "Github hiring tests repository management",
	}

	cmd.AddCommand(NewHiringSendCmd(ctx))
	cmd.AddCommand(NewHiringUnseat(ctx))

	return cmd
}
