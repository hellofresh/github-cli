package cmd

import (
	"context"
	"os"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/hellofresh/github-cli/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	cfgFile      string
	globalConfig *config.Spec
	githubClient *github.Client
	version      string
	token        string

	// RootCmd is our main command
	RootCmd = &cobra.Command{
		Use:   "github-cli [--config] [--token]",
		Short: "HF Github is a cli tool to manage your github repositories",
		Long: `A simple CLI tool to help you manage your github repositories.
		Complete documentation is available at http://github.com/hellofresh/github-cli`,
	}

	// Repo commands
	repoCmd = &cobra.Command{
		Use:     "repo",
		Aliases: []string{"re"},
	}

	createRepoCmd = &cobra.Command{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "Creates a new github repository",
		Long:    `Creates a new github repository based on the rules defined on your .github.toml`,
		Run:     RunCreateRepo,
	}

	// Hiring tests commands
	testsCmd = &cobra.Command{
		Use:     "test",
		Aliases: []string{"te"},
	}

	createTestCmd = &cobra.Command{
		Use:     "create",
		Aliases: []string{"ct"},
		Short:   "Creates a new hellofresh hiring test",
		Long:    `Creates a new hellofresh hiring test based on the rules defined on your .github.toml`,
		Run:     RunCreateTestRepo,
	}

	// Unseat commands
	unseatCmd = &cobra.Command{
		Use:     "unseat",
		Aliases: []string{"un"},
		Short:   "Removes external collaborators from repositories",
		Long:    `Removes external (people not in the organization) collaborators from repositories`,
		Run:     RunUnseat,
	}
)

func init() {
	var err error

	globalConfig, err = config.Load(cfgFile)
	if err != nil {
		log.WithError(err).Error("Could not load the configurations")
	}

	log.SetOutput(os.Stdout)

	lvl, err := log.ParseLevel(globalConfig.LogLevel)
	if err != nil {
		log.WithError(err).Error("Couldn't parse the log level")
	}

	// Only log the warning severity or above.
	log.SetLevel(lvl)

	if token != "" {
		globalConfig.Github.Token = token
		globalConfig.GithubTestOrg.Token = token
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: globalConfig.Github.Token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	githubClient = github.NewClient(tc)
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is /etc/github/.github.toml)")
	RootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "config file (default is /etc/github/.github.toml)")

	// Aggregates Root commands
	RootCmd.AddCommand(repoCmd)
	RootCmd.AddCommand(testsCmd)
	RootCmd.AddCommand(unseatCmd)
}

func checkEmpty(value interface{}, msg string) {
	switch v := value.(type) {
	case error:
		if v != nil {
			color.Red(v.Error())
			os.Exit(1)
		}
	case string:
		if v == "" {
			color.Red(msg)
			os.Exit(1)
		}
	}
}
