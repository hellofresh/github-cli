package cmd

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/hellofresh/github-cli/pkg/config"
	"github.com/hellofresh/github-cli/pkg/formatter"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	cfgFile      string
	globalConfig *config.Spec
	githubClient *github.Client
	token        string
	verbose      bool

	// RootCmd is our main command
	RootCmd = &cobra.Command{
		Use:   "github-cli [--config] [--token]",
		Short: "HF Github is a cli tool to manage your github repositories",
		Long: `A simple CLI tool to help you manage your github repositories.
		Complete documentation is available at http://github.com/hellofresh/github-cli`,
	}
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.github.toml)")
	RootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "optional, github token for authentication (default in $HOME/.github.toml)")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Make the operation more talkative")

	// Aggregates Root commands
	RootCmd.AddCommand(NewRepoCmd())
	RootCmd.AddCommand(NewHiringCmd())
	RootCmd.AddCommand(NewVersionCmd())

	log.SetFormatter(&formatter.CliFormatter{})
}

func setupConnection(cmd *cobra.Command, args []string) error {
	var err error

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	globalConfig, err = config.Load(cfgFile)
	if err != nil {
		return errors.Wrap(err, "Could not load the configurations")
	}

	if token != "" {
		globalConfig.Github.Token = token
		globalConfig.GithubTestOrg.Token = token
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: globalConfig.Github.Token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	githubClient = github.NewClient(tc)

	return nil
}
