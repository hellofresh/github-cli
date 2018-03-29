package cmd

import (
	"context"

	"github.com/hellofresh/github-cli/pkg/config"
	"github.com/hellofresh/github-cli/pkg/github"
	"github.com/hellofresh/github-cli/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type (
	// RootOptions represents the ahoy global options
	RootOptions struct {
		configFile string
		token      string
		org        string
		verbose    bool
	}
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	opts := RootOptions{}

	cmd := cobra.Command{
		Use:   "github-cli [--config] [--token]",
		Short: "HF Github is a cli tool to manage your github repositories",
		PersistentPreRun: func(ccmd *cobra.Command, args []string) {
			if opts.verbose {
				log.WithContext(context.Background()).SetLevel(logrus.DebugLevel)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.configFile, "config", "c", "", "config file (default is $HOME/.github.toml)")
	cmd.PersistentFlags().StringVarP(&opts.token, "token", "t", "", "optional, github token for authentication (default in $HOME/.github.toml)")
	cmd.PersistentFlags().BoolVarP(&opts.verbose, "verbose", "v", false, "Make the operation more talkative")
	cmd.PersistentFlags().StringVarP(&opts.org, "organization", "o", "", "Github's organization")

	ctx := log.NewContext(context.Background())
	ctx, err := config.NewContext(ctx, opts.configFile)
	if err != nil {
		log.WithContext(ctx).WithError(err).Fatal("Could not load configuration file")
	}

	cfg := config.WithContext(ctx)
	if opts.token != "" {
		cfg.Github.Token = opts.token
		cfg.GithubTestOrg.Token = opts.token
	}

	if opts.org != "" {
		cfg.Github.Organization = opts.org
		cfg.GithubTestOrg.Organization = opts.org
	}
	ctx = config.OverrideConfig(ctx, cfg)

	ctx, err = github.NewContext(ctx, cfg.Github.Token)
	if err != nil {
		log.WithContext(ctx).WithError(err).Fatal("Could not create the kube client")
	}

	// Aggregates Root commands
	cmd.AddCommand(NewRepoCmd(ctx))
	cmd.AddCommand(NewHiringCmd(ctx))
	cmd.AddCommand(NewVersionCmd(ctx))
	cmd.AddCommand(NewUpdateCmd(ctx))

	return &cmd
}
