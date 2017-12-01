package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version = "0.0.0-dev"

	versionCmd = &cobra.Command{
		Use:     "version",
		Short:   "Print the version information",
		Aliases: []string{"v"},
		Run:     RunVersion,
	}
)

// RunVersion runs the command to print the current version
func RunVersion(cmd *cobra.Command, args []string) {
	color.Green("github-cli %s", version)
}
