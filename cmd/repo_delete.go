package cmd

import (
	"context"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type (
	// DeleteRepoOpts are the flags for the delete repo command
	DeleteRepoOpts struct {
		Org  string
		Name string
	}
)

// NewDeleteRepoCmd creates a new delete repo command
func NewDeleteRepoCmd() *cobra.Command {
	opts := &DeleteRepoOpts{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Deletes a github repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunDeleteRepo(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "The name of the repository")
	cmd.Flags().StringVarP(&opts.Org, "organization", "o", "", "Github's organization")

	return cmd
}

// RunDeleteRepo runs the command to delete a repository
func RunDeleteRepo(opts *DeleteRepoOpts) error {
	org := opts.Org
	if org == "" {
		org = globalConfig.Github.Organization
	}
	if org == "" {
		return errors.New("Please provide an organization")
	}

	_, err := githubClient.Repositories.Delete(context.Background(), org, opts.Name)
	if err != nil {
		return errors.Wrap(err, "Could not delete repository")
	}

	log.Info("Repository %s deleted!", opts.Name)

	return nil
}
