package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	"github.com/deiwin/interact"
	"github.com/hellofresh/github-cli/pkg/config"
	"github.com/hellofresh/github-cli/pkg/repo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// RunCreateTestRepo runs the command to create a new hiring test repository
func RunCreateTestRepo(cmd *cobra.Command, args []string) {
	var err error
	actor := interact.NewActor(os.Stdin, os.Stdout)
	org, err := actor.PromptOptional("Please enter the org name", globalConfig.Github.Organization, checkNotEmpty)
	if err != nil {
		log.Fatal(err)
	}

	candidate, err := actor.Prompt("GitHub username of the candidate", checkNotEmpty)
	if err != nil {
		log.Fatal(err)
	}

	testRepo, err := actor.Prompt("Name of the repo with the test", checkNotEmpty)
	if err != nil {
		log.Fatal(err)
	}

	target := fmt.Sprintf("%s-%s", candidate, testRepo)
	opts := &repo.HelloFreshRepoOpt{
		Name:    target,
		Org:     org,
		Private: true,
		Collaborators: &repo.CollaboratorsRule{
			Enabled: true,
			Collaborators: []*config.Collaborator{
				&config.Collaborator{
					Username:   candidate,
					Permission: "push",
				},
			},
		},
	}
	creator := repo.NewGithub(githubClient)
	err = creator.Create(opts)
	if err != nil {
		log.WithError(err).Fatal("Could not create github repo for candidate")
	}

	if globalConfig.PublicKeyPath == "" {
		user, _ := user.Current()
		globalConfig.PublicKeyPath = fmt.Sprintf("%s/.ssh/id_rsa", user.HomeDir)
	}

	sshKey, err := ioutil.ReadFile(globalConfig.PublicKeyPath)
	if err != nil {
		log.WithError(err).Fatal("Error reading public key")
	}

	authMethod, err := ssh.NewPublicKeys("git", []byte(sshKey), "")
	if err != nil {
		log.WithError(err).Fatal("Error when creating public keys")
	}

	log.Info("Cloning repository...")
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		Auth:     authMethod,
		Progress: os.Stdout,
		URL:      fmt.Sprintf("git@github.com:%s/%s", org, testRepo),
	})
	if err != nil {
		log.WithError(err).Fatal("Error cloning to repository")
	}

	log.Info("Changing remote...")
	remote, err := r.Remote(git.DefaultRemoteName)
	if err != nil {
		log.WithError(err).Fatal("Error changing remote for repository")
	}

	log.Info("Pushing changes...")
	remote.Config().URLs = []string{fmt.Sprintf("git@github.com:%s/%s", org, target)}
	err = remote.Push(&git.PushOptions{
		RemoteName: git.DefaultRemoteName,
		Progress:   os.Stdout,
	})
	if err != nil {
		log.WithError(err).Fatal("Error pushing to repository")
	}

	log.Infof("Done! Test for %s is created", candidate)
}
