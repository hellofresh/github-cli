package cmd

import (
	"fmt"
	"os"
	"os/user"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/hellofresh/github-cli/pkg/config"
	"github.com/hellofresh/github-cli/pkg/repo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type (
	// CreateTestRepoFlags are the flags for the create a hiring test repository command
	CreateTestRepoFlags struct {
		Org       string
		Candidate string
		TestRepo  string
	}
)

var createTestRepoFlags CreateTestRepoFlags

func init() {
	testsCmd.AddCommand(createTestCmd)

	createTestCmd.Flags().StringVarP(&createTestRepoFlags.TestRepo, "test-repo", "r", "", "The name of the test repository to clone from")
	createTestCmd.Flags().StringVarP(&createTestRepoFlags.Candidate, "username", "u", "", "The github's name of the candidate")
	createTestCmd.Flags().StringVarP(&createTestRepoFlags.Org, "organization", "o", "", "Github's organization")
}

// RunCreateTestRepo runs the command to create a new hiring test repository
func RunCreateTestRepo(cmd *cobra.Command, args []string) {
	var err error

	org := createTestRepoFlags.Org
	if org == "" {
		org = globalConfig.GithubTestOrg.Organization
	}
	checkEmpty(org, "Please provide an organization")

	candidate := createTestRepoFlags.Candidate
	checkEmpty(org, "Please provide a candidate username")

	testRepo := createTestRepoFlags.TestRepo
	checkEmpty(org, "Please provide a test repository")

	target := fmt.Sprintf("%s-%s", candidate, testRepo)

	creator := repo.NewGithub(githubClient)

	color.White("Creating repository...")
	_, err = creator.CreateRepo(org, &github.Repository{
		Name:      github.String(target),
		Private:   github.Bool(true),
		HasIssues: github.Bool(false),
		HasPages:  github.Bool(false),
		HasWiki:   github.Bool(false),
	})
	checkEmpty(err, "Could not create github repo for candidate")

	color.White("Adding collaborators to repository...")
	opts := []*config.Collaborator{
		&config.Collaborator{
			Username:   candidate,
			Permission: "push",
		},
	}
	err = creator.AddCollaborators(target, org, opts)
	checkEmpty(errors.Wrap(err, "could not add collaborators to repository"), "")

	if globalConfig.PublicKeyPath == "" {
		user, _ := user.Current()
		globalConfig.PublicKeyPath = fmt.Sprintf("%s/.ssh/id_rsa", user.HomeDir)
	}

	color.White("Cloning repository...")
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		Progress: os.Stdout,
		URL:      fmt.Sprintf("https://%s@github.com/%s/%s", globalConfig.GithubTestOrg.Token, org, testRepo),
	})
	checkEmpty(err, "Error cloning to repository")

	color.White("Changing remote...")
	remote, err := r.Remote(git.DefaultRemoteName)
	checkEmpty(err, "Error changing remote for repository")

	color.White("Pushing changes...")
	remote.Config().URLs = []string{fmt.Sprintf("git@github.com:%s/%s", org, target)}
	err = remote.Push(&git.PushOptions{
		RemoteName: git.DefaultRemoteName,
		Progress:   os.Stdout,
	})
	checkEmpty(err, "Error pushing to repository")

	color.Green("Done! Hiring test for %s is created", candidate)
}
