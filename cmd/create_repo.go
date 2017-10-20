package cmd

import (
	"errors"
	"os"

	"github.com/deiwin/interact"
	"github.com/hellofresh/github-cli/pkg/pullapprove"
	"github.com/hellofresh/github-cli/pkg/repo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type (
	// CreateRepoFlags are the flags for the create repository command
	CreateRepoFlags struct {
		Private              bool
		HasPullApprove       bool
		HasTeams             bool
		HasLabels            bool
		HasDefaultLabels     bool
		HasWebhooks          bool
		HasBranchProtections bool
	}
)

var (
	createRepoFlags CreateRepoFlags
	checkNotEmpty   = func(input string) error {
		// note that the inputs provided to these checks are already trimmed
		if input == "" {
			return errors.New("input should not be empty")
		}
		return nil
	}
)

// RunCreateRepo runs the command to create a new repository
func RunCreateRepo(cmd *cobra.Command, args []string) {
	var err error

	actor := interact.NewActor(os.Stdin, os.Stdout)
	repoName, err := actor.Prompt("Please enter the repository name", checkNotEmpty)
	if err != nil {
		log.Fatal(err)
	}

	org, err := actor.PromptOptional("Please enter the org name", globalConfig.Github.Organization, checkNotEmpty)
	if err != nil {
		log.Fatal(err)
	}

	opts := &repo.HelloFreshRepoOpt{
		Name:    repoName,
		Org:     org,
		Private: createRepoFlags.Private,
		PullApprove: &repo.PullApproveRule{
			Enabled:             createRepoFlags.HasPullApprove,
			Filename:            globalConfig.PullApprove.Filename,
			ProtectedBranchName: globalConfig.PullApprove.ProtectedBranchName,
			Client:              pullapprove.New(globalConfig.PullApprove.Token),
		},
		Labels: &repo.LabelsRule{
			Enabled:             createRepoFlags.HasLabels,
			RemoveDefaultLabels: globalConfig.Github.RemoveDefaultLabels,
			Labels:              globalConfig.Github.Labels,
		},
		Teams: &repo.TeamsRule{
			Enabled: createRepoFlags.HasTeams,
			Teams:   globalConfig.Github.Teams,
		},
		Webhooks: &repo.WebhooksRule{
			Enabled:  createRepoFlags.HasWebhooks,
			Webhooks: globalConfig.Github.Webhooks,
		},
		BranchProtections: &repo.BranchProtectionsRule{
			Enabled:     createRepoFlags.HasBranchProtections,
			Protections: globalConfig.Github.Protections,
		},
	}

	creator := repo.NewGithub(githubClient)
	err = creator.Create(opts)
	if err != nil {
		log.WithError(err).Fatal("Could not create github repo")
	}
}
