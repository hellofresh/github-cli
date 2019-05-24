package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/hashicorp/errwrap"
	"github.com/hellofresh/github-cli/pkg/config"
	gh "github.com/hellofresh/github-cli/pkg/github"
	"github.com/hellofresh/github-cli/pkg/log"
	"github.com/hellofresh/github-cli/pkg/repo"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type (
	// HiringSendOpts are the flags for the send a hiring test command
	HiringSendOpts struct{}
)

// NewHiringSendCmd creates a new send hiring test command
func NewHiringSendCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [username] [repo] [branch]",
		Short: "Creates a new HelloFresh hiring test",
		Long:  `Creates a new HelloFresh hiring test based on the rules defined on your .github.toml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			branch := plumbing.Master
			if len(args) > 2 {
				branch = plumbing.ReferenceName("refs/heads/" + args[2])
			}
			return RunCreateTestRepo(ctx, args[0], args[1], branch)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 || args[0] == "" {
				return errors.New("Please provide a github username for the candidate")
			}

			if len(args) < 2 || args[1] == "" {
				return errors.New("Please provide which repository test")
			}

			return nil
		},
	}

	return cmd
}

// RunCreateTestRepo runs the command to create a new hiring test repository
func RunCreateTestRepo(ctx context.Context, candidate string, testRepo string, reference plumbing.ReferenceName) error {
	var err error

	logger := log.WithContext(ctx)
	cfg := config.WithContext(ctx)
	githubClient := gh.WithContext(ctx)
	if githubClient == nil {
		return errors.New("failed to get github client")
	}

	org := cfg.GithubTestOrg.Organization
	if org == "" {
		return errors.New("please provide an organization")
	}

	target := fmt.Sprintf("%s-%s", candidate, testRepo)

	creator := repo.NewGithub(githubClient)

	logger.Info("Creating repository...")
	_, err = creator.CreateRepo(ctx, org, &github.Repository{
		Name:      github.String(target),
		Private:   github.Bool(true),
		HasIssues: github.Bool(false),
		HasPages:  github.Bool(false),
		HasWiki:   github.Bool(false),
	})
	if err != nil {
		return errwrap.Wrapf("could not create github repo for candidate: {{err}}", err)
	}

	logger.Info("Adding collaborators to repository...")
	collaboratorsOpts := []*config.Collaborator{
		&config.Collaborator{
			Username:   candidate,
			Permission: "push",
		},
	}
	err = creator.AddCollaborators(ctx, target, org, collaboratorsOpts)
	if err != nil {
		return errwrap.Wrapf("could not add collaborators to repository: {{err}}", err)
	}

	logger.Info("Cloning repository...")
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		Progress:      os.Stdout,
		URL:           fmt.Sprintf("https://%s@github.com/%s/%s", cfg.GithubTestOrg.Token, org, testRepo),
		ReferenceName: reference,
	})
	if err != nil {
		return errwrap.Wrapf("error cloning to repository: {{err}}", err)
	}

	logger.Info("Changing remote...")
	remote, err := r.Remote(git.DefaultRemoteName)
	if err != nil {
		return errwrap.Wrapf("error changing remote for repository: {{err}}", err)
	}

	logger.Info("Pushing changes...")
	remote.Config().URLs = []string{fmt.Sprintf("https://%s@github.com/%s/%s", cfg.GithubTestOrg.Token, org, target)}
	err = remote.Push(&git.PushOptions{
		RemoteName: git.DefaultRemoteName,
		Progress:   os.Stdout,
	})
	if err != nil {
		return errwrap.Wrapf("error pushing to repository: {{err}}", err)
	}

	logger.Infof("Done! Hiring test for %s is created", candidate)

	return nil
}
