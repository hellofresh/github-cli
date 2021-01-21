package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/hellofresh/github-cli/pkg/config"
	gh "github.com/hellofresh/github-cli/pkg/github"
	"github.com/hellofresh/github-cli/pkg/log"
	"github.com/hellofresh/github-cli/pkg/pullapprove"
	"github.com/hellofresh/github-cli/pkg/repo"
	"github.com/hellofresh/github-cli/pkg/zappr"
)

// CreateRepoOptions are the flags for the create repository command
type CreateRepoOptions struct {
	Description               string
	Private                   bool
	HasPullApprove            bool
	HasZappr                  bool
	UseZapprGithubCredentials bool
	HasTeams                  bool
	HasCollaborators          bool
	HasLabels                 bool
	HasDefaultLabels          bool
	HasWebhooks               bool
	HasBranchProtections      bool
	HasIssues                 bool
	HasWiki                   bool
	HasPages                  bool
}

// NewCreateRepoCmd creates a new create repo command
func NewCreateRepoCmd(ctx context.Context) *cobra.Command {
	opts := &CreateRepoOptions{}

	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Creates a new github repository",
		Long:  `Creates a new github repository based on the rules defined on your .github.toml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunCreateRepo(ctx, args[0], opts)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 || args[0] == "" {
				return errors.New("please provide a repository name")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "The repository's description")
	cmd.Flags().BoolVar(&opts.Private, "private", true, "Is the repository private?")

	cmd.Flags().BoolVar(&opts.HasIssues, "has-issues", true, "Enables issue pages")
	cmd.Flags().BoolVar(&opts.HasWiki, "has-wiki", false, "Enables wiki pages?")
	cmd.Flags().BoolVar(&opts.HasPages, "has-pages", false, "Enables github pages?")
	cmd.Flags().BoolVar(&opts.HasPullApprove, "has-pullapprove", false, "Enables pull approve")
	cmd.Flags().BoolVar(&opts.HasZappr, "has-zappr", true, "Enables Zappr")
	cmd.Flags().BoolVar(&opts.HasTeams, "has-teams", true, "Enable teams")
	cmd.Flags().BoolVar(&opts.HasLabels, "has-labels", true, "Enable labels")
	cmd.Flags().BoolVar(&opts.HasDefaultLabels, "rm-default-labels", true, "Removes the default github labels")
	cmd.Flags().BoolVar(&opts.HasWebhooks, "has-webhooks", false, "Enables webhooks configurations")
	cmd.Flags().BoolVar(&opts.HasBranchProtections, "has-branch-protections", true, "Enables branch protections")
	cmd.Flags().BoolVar(&opts.UseZapprGithubCredentials, "use-zappr-credentials", true, "Enables authenticating to Github as Zapps App")

	return cmd
}

// RunCreateRepo runs the command to create a new repository
func RunCreateRepo(ctx context.Context, repoName string, opts *CreateRepoOptions) error {
	wg, ctx := errgroup.WithContext(ctx)
	logger := log.WithContext(ctx)
	cfg := config.WithContext(ctx)
	githubClient := gh.WithContext(ctx)
	if githubClient == nil {
		return errors.New("failed to get github client")
	}

	org := cfg.Github.Organization
	if org == "" {
		return errors.New("please provide an organization")
	}

	logger.Debug("Create options:")
	logger.Debugf("\tConfigure PullApprove? %s", strconv.FormatBool(opts.HasPullApprove))
	logger.Debugf("\tConfigure Zappr? %s", strconv.FormatBool(opts.HasZappr))
	logger.Debugf("\tConfigure GitHub teams? %s", strconv.FormatBool(opts.HasTeams))
	logger.Debugf("\tConfigure collaborators? %s", strconv.FormatBool(opts.HasCollaborators))
	logger.Debugf("\tAdd labels to repository? %s", strconv.FormatBool(opts.HasLabels))
	logger.Debugf("\tAdd webhooks to repository? %s", strconv.FormatBool(opts.HasWebhooks))
	logger.Debugf("\tConfigure branch protection? %s", strconv.FormatBool(opts.HasBranchProtections))
	logger.Debugf("\tAuthenticate to Github as Zappr? %s", strconv.FormatBool(opts.UseZapprGithubCredentials))

	description := opts.Description
	githubOpts := &repo.GithubRepoOpts{
		PullApprove: &repo.PullApproveOpts{
			Filename:            cfg.PullApprove.Filename,
			ProtectedBranchName: cfg.PullApprove.ProtectedBranchName,
			Client:              pullapprove.New(cfg.PullApprove.Token),
		},
		Labels: &repo.LabelsOpts{
			RemoveDefaultLabels: cfg.Github.RemoveDefaultLabels,
			Labels:              cfg.Github.Labels,
		},
		Teams:             cfg.Github.Teams,
		Webhooks:          cfg.Github.Webhooks,
		BranchProtections: cfg.Github.Protections,
	}

	creator := repo.NewGithub(githubClient)

	logger.Infof("Creating repository %s/%s...", org, repoName)
	ghRepo, err := creator.CreateRepo(ctx, org, &github.Repository{
		Name:        github.String(repoName),
		Description: github.String(description),
		Private:     github.Bool(opts.Private),
		HasIssues:   github.Bool(opts.HasIssues),
		HasWiki:     github.Bool(opts.HasWiki),
		HasPages:    github.Bool(opts.HasPages),
		AutoInit:    github.Bool(true),
	})
	if err == repo.ErrRepositoryAlreadyExists {
		logger.Info("Repository already exists. Trying to normalize it...")
	} else if err != nil {
		return fmt.Errorf("could not create repository: %w", err)
	}

	if opts.HasPullApprove {
		wg.Go(func() error {
			logger.Info("Adding pull approve...")
			err = creator.AddPullApprove(ctx, repoName, org, githubOpts.PullApprove)
			if err == repo.ErrPullApproveFileAlreadyExists {
				logger.Debug("Pull approve already exists, moving on...")
			} else if err != nil {
				return fmt.Errorf("could not add pull approve: %w", err)
			}

			return nil
		})
	}

	if opts.HasZappr {
		wg.Go(func() error {
			logger.Info("Adding Zappr...")

			if ghRepo == nil {
				logger.Debug("Fetching repo details from Github")
				ghRepo, _, err = githubClient.Repositories.Get(ctx, org, repoName)

				if err != nil {
					return fmt.Errorf("information required to enable zappr on github repo was not found: %w", err)
				}
			}

			var zapprClient zappr.Client
			zapprClient = zappr.New(cfg.Zappr.URL, cfg.Github.Token, nil)

			if cfg.Zappr.UseZapprGithubCredentials {
				logger.Debug("Retrieving token for zappr github app from zappr")
				err = zapprClient.ImpersonateGitHubApp()
				if err != nil {
					if errors.Is(err, zappr.ErrZapprUnauthorized) {
						return fmt.Errorf("could not retrieve token representing github zappr app from zappr. it seems you have not logged in to zappr, if you have, please logout from zappr, log back in and try again: %w", err)
					}

					return fmt.Errorf("could not retrieve token representing github zappr app from zappr: %w", err)
				}
			}

			err = zapprClient.Enable(int(*ghRepo.ID))
			if errors.Is(err, zappr.ErrZapprAlreadyEnabled) {
				logger.Debug("Zappr already enabled, moving on...")
			} else if err != nil {
				return fmt.Errorf("could not enable zappr: %w", err)
			}

			return nil
		})
	}

	if opts.HasTeams {
		wg.Go(func() error {
			logger.Info("Adding teams to repository...")
			err = creator.AddTeamsToRepo(ctx, repoName, org, githubOpts.Teams)
			if err != nil {
				return fmt.Errorf("could not add teams to repository: %w", err)
			}

			return nil
		})
	}

	if opts.HasCollaborators {
		wg.Go(func() error {
			logger.Info("Adding collaborators to repository...")
			if err = creator.AddCollaborators(ctx, repoName, org, githubOpts.Collaborators); err != nil {
				return fmt.Errorf("could not add collaborators to repository: %w", err)
			}

			return nil
		})
	}

	if opts.HasLabels {
		wg.Go(func() error {
			logger.Info("Adding labels to repository...")
			if err := creator.AddLabelsToRepo(ctx, repoName, org, githubOpts.Labels); err != nil {
				if strings.Contains(err.Error(), repo.ErrLabeAlreadyExists.Error()) {
					logger.Debug("Labels already exists, moving on...")
					return nil
				}

				return fmt.Errorf("could not add labels to repository: %w", err)
			}

			return nil
		})
	}

	if opts.HasWebhooks {
		wg.Go(func() error {
			logger.Info("Adding webhooks to repository...")
			if err := creator.AddWebhooksToRepo(ctx, repoName, org, githubOpts.Webhooks); err != nil {
				if strings.Contains(err.Error(), repo.ErrWebhookAlreadyExist.Error()) {
					logger.Debug("Webhook already exists, moving on...")
					return nil
				}
				return fmt.Errorf("could not add webhooks to repository: %w", err)
			}

			return nil
		})
	}

	if opts.HasBranchProtections {
		wg.Go(func() error {
			logger.Info("Adding branch protections to repository...")
			if err = creator.AddBranchProtections(ctx, repoName, org, githubOpts.BranchProtections); err != nil {
				return fmt.Errorf("could not add branch protections to repository: %w", err)
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return err
	}

	if ghRepo != nil {
		logger.Infof("Repository created! \n Here is how to access it %s", ghRepo.GetGitURL())
	} else {
		logger.Info("Repository normalized!")
	}

	return nil
}
