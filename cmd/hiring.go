package cmd

import "github.com/spf13/cobra"

// NewHiringCmd aggregates the hiring comamnds
func NewHiringCmd() *cobra.Command {
	// Repo commands
	cmd := &cobra.Command{
		Use:   "hiring",
		Short: "Github hiring tests repository management",
	}

	cmd.AddCommand(NewHiringSendCmd())
	cmd.AddCommand(NewHiringUnseat())

	return cmd
}
