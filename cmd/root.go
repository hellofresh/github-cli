package cmd

import (
	"os"

	"github.com/hellofresh/github-cli/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cfgFile      string
	globalConfig *config.Spec
	version      string
	// RootCmd is our main command
	RootCmd = &cobra.Command{
		Use:   "github-cli",
		Short: "HF Github is a cli tool to manage your github repositories",
		Long: `A simple CLI tool to help you manage your github repositories.
		Complete documentation is available at http://github.com/hellofresh/github-cli`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error

			globalConfig, err = config.Load(cfgFile)
			if err != nil {
				log.WithError(err).Error("Could not load the configurations")
			}

			// Log as JSON instead of the default ASCII formatter.
			// log.SetFormatter(&log.JSONFormatter{})

			// Output to stdout instead of the default stderr
			// Can be any io.Writer, see below for File example
			log.SetOutput(os.Stdout)

			lvl, err := log.ParseLevel(globalConfig.LogLevel)
			if err != nil {
				log.WithError(err).Error("Couldn't parse the log level")
				os.Exit(1)
			}

			// Only log the warning severity or above.
			log.SetLevel(lvl)
		},
	}

	createRepoCmd = &cobra.Command{
		Use:   "create",
		Short: "Creates a new github repository",
		Long:  `Creates a new github repository based on the rules defined on your .github.toml`,
		Run:   RunCreate,
	}
)

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/github/.github.toml)")

	createRepoCmd.Flags().BoolVar(&createRepoFlags.Private, "private", true, "Is the repository private?")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasPullApprove, "add-pullapprove", true, "Enables pull approve")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasTeams, "add-teams", true, "Enable teams")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasLabels, "add-labels", true, "Enable labels")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasDefaultLabels, "add-default-labels", true, "Removes the default github labels")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasWebhooks, "add-webhooks", true, "Enables webhooks configurations")
	createRepoCmd.Flags().BoolVar(&createRepoFlags.HasBranchProtections, "add-branch-protections", true, "Enables branch protections")
	RootCmd.AddCommand(createRepoCmd)
}
