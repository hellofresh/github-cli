package cmd

import (
	"context"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type (
	// DeleteRepoFlags are the flags for the delete repo command
	DeleteRepoFlags struct {
		Org  string
		Name string
	}
)

var (
	deleteRepoFlags DeleteRepoFlags

	// Delete repo command
	deleteRepoCmd = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"del"},
		Short:   "Deletes a github repository",
		Run:     RunDeleteRepo,
	}
)

func init() {
	repoCmd.AddCommand(deleteRepoCmd)

	deleteRepoCmd.Flags().StringVarP(&deleteRepoFlags.Name, "name", "n", "", "The name of the repository")
	deleteRepoCmd.Flags().StringVarP(&deleteRepoFlags.Org, "organization", "o", "", "Github's organization")
}

// RunDeleteRepo runs the command to delete a repository
func RunDeleteRepo(cmd *cobra.Command, args []string) {
	org := deleteRepoFlags.Org
	if org == "" {
		org = globalConfig.Github.Organization
	}
	checkEmpty(org, "Please provide an organization")

	_, err := githubClient.Repositories.Delete(context.Background(), org, deleteRepoFlags.Name)
	checkEmpty(err, "Could not delete repository")

	color.Green("Repository %s deleted!", deleteRepoFlags.Name)
}
