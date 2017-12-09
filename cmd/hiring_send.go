package cmd

import (
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/hellofresh/github-cli/pkg/config"
	"github.com/hellofresh/github-cli/pkg/repo"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type (
	// HiringSendOpts are the flags for the send a hiring test command
	HiringSendOpts struct {
		Org string
	}
)

// NewHiringSendCmd creates a new send hiring test command
func NewHiringSendCmd() *cobra.Command {
	opts := &HiringSendOpts{}
	cmd := &cobra.Command{
		Use:     "send [username] [repo]",
		Short:   "Creates a new hellofresh hiring test",
		Long:    `Creates a new hellofresh hiring test based on the rules defined on your .github.toml`,
		PreRunE: setupConnection,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunCreateTestRepo(args[0], args[1], opts)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if args[0] == "" {
				return errors.New("Please provide a github username for the candidate")
			}

			if args[1] == "" {
				return errors.New("Please provide which repository test")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Org, "organization", "o", "", "Github's organization")

	return cmd
}

// RunCreateTestRepo runs the command to create a new hiring test repository
func RunCreateTestRepo(candidate string, testRepo string, opts *HiringSendOpts) error {
	var err error

	org := opts.Org
	if org == "" {
		org = globalConfig.GithubTestOrg.Organization
	}
	if org == "" {
		return errors.New("Please provide an organization")
	}

	target := fmt.Sprintf("%s-%s", candidate, testRepo)

	creator := repo.NewGithub(githubClient)

	log.Info("Creating repository...")
	_, err = creator.CreateRepo(org, &github.Repository{
		Name:      github.String(target),
		Private:   github.Bool(true),
		HasIssues: github.Bool(false),
		HasPages:  github.Bool(false),
		HasWiki:   github.Bool(false),
	})
	if err != nil {
		return errors.Wrap(err, "Could not create github repo for candidate")
	}

	log.Info("Adding collaborators to repository...")
	collaboratorsOpts := []*config.Collaborator{
		&config.Collaborator{
			Username:   candidate,
			Permission: "push",
		},
	}
	err = creator.AddCollaborators(target, org, collaboratorsOpts)
	if err != nil {
		return errors.Wrap(err, "Could not add collaborators to repository")
	}

	log.Info("Cloning repository...")
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		Progress: os.Stdout,
		URL:      fmt.Sprintf("https://%s@github.com/%s/%s", globalConfig.GithubTestOrg.Token, org, testRepo),
	})
	if err != nil {
		return errors.Wrap(err, "Error cloning to repository")
	}

	log.Info("Changing remote...")
	remote, err := r.Remote(git.DefaultRemoteName)
	if err != nil {
		return errors.Wrap(err, "Error changing remote for repository")
	}

	log.Info("Pushing changes...")
	remote.Config().URLs = []string{fmt.Sprintf("https://%s@github.com/%s/%s", globalConfig.GithubTestOrg.Token, org, target)}
	err = remote.Push(&git.PushOptions{
		RemoteName: git.DefaultRemoteName,
		Progress:   os.Stdout,
	})
	if err != nil {
		return errors.Wrap(err, "Error pushing to repository")
	}

	log.Infof("Done! Hiring test for %s is created", candidate)

	return nil
}
