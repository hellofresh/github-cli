package cmd

import (
	"errors"
	"os"

	"github.com/deiwin/interact"
	"github.com/hellofresh/github-cli/pkg/pullapprove"

	"github.com/hellofresh/github-cli/pkg/github"
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

func RunCreate(cmd *cobra.Command, args []string) {
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

	opts := &github.HelloFreshRepoOpt{
		Token:   globalConfig.Github.Token,
		Private: createRepoFlags.Private,
		PullApprove: &github.PullApproveRule{
			Enabled:             createRepoFlags.HasPullApprove,
			Filename:            globalConfig.PullApprove.Filename,
			ProtectedBranchName: globalConfig.PullApprove.ProtectedBranchName,
			Client:              pullapprove.New(globalConfig.PullApprove.Token),
		},
		Labels: &github.LabelsRule{
			Enabled:             createRepoFlags.HasLabels,
			RemoveDefaultLabels: globalConfig.Github.RemoveDefaultLabels,
			Labels:              globalConfig.Github.Labels,
		},
		Teams: &github.TeamsRule{
			Enabled: createRepoFlags.HasTeams,
			Teams:   globalConfig.Github.Teams,
		},
		Webhooks: &github.WebhooksRule{
			Enabled:  createRepoFlags.HasWebhooks,
			Webhooks: globalConfig.Github.Webhooks,
		},
		BranchProtections: &github.BranchProtectionsRule{
			Enabled:     createRepoFlags.HasBranchProtections,
			Protections: globalConfig.Github.Protections,
		},
	}

	creator, err := github.NewHelloFreshRepoCreator(repoName, org, opts)
	if err != nil {
		log.WithError(err).Error("Could not create github client")
		os.Exit(1)
	}

	err = creator.Create()
	if err != nil {
		log.WithError(err).Error("Could not create github repo")
		os.Exit(1)
	}
}
