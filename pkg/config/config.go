package config

import (
	"context"
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/hellofresh/github-cli/pkg/log"
)

type (
	configKeyType int

	// Spec represents the global app configuration
	Spec struct {
		Github        Github
		GithubTestOrg Github
		PullApprove   PullApprove
		Zappr         Zappr
	}

	// PullApprove represents teh Pull Approve configurations
	PullApprove struct {
		Token               string
		Filename            string
		ProtectedBranchName string
	}

	// Zappr represents Zappr configurations
	Zappr struct {
		URL                       string
		UseZapprGithubCredentials bool
	}

	// Github represents the github configurations
	Github struct {
		Organization  string
		Token         string
		Teams         []*Team
		Collaborators []*Collaborator
		Labels        []*Label
		Webhooks      []*Webhook
		Protections   BranchProtections
		// RemoveDefaultLabels Remove GitHub's default labels?
		RemoveDefaultLabels bool
	}

	// BranchProtections represents github's branch protections
	BranchProtections map[string][]string

	// Team represents a github team
	Team struct {
		ID         int
		Permission string
	}

	// Collaborator represents a github collaborator
	Collaborator struct {
		Username   string
		Permission string
	}

	// Label represents a github label
	Label struct {
		Name  string
		Color string
	}

	// Webhook represents a github webhook
	Webhook struct {
		Type   string
		Config map[string]interface{}
	}
)

const configKey configKeyType = iota

// NewContext loads a configuration file into the Spec struct
func NewContext(ctx context.Context, configFile string) (context.Context, error) {
	logger := log.WithContext(ctx)

	if configFile != "" {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return ctx, fmt.Errorf("invalid configuration file provided %s", configFile)
		}
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {
		homeDir, err := homedir.Dir()
		if err != nil {
			return ctx, err
		}

		viper.SetConfigName(".github")
		viper.AddConfigPath(".")
		viper.AddConfigPath(homeDir)
	}

	logger.Debugf("Reading config from %s...", viper.ConfigFileUsed())

	viper.SetDefault("github.token", os.Getenv("GITHUB_TOKEN"))
	viper.SetDefault("githubtestorg.token", os.Getenv("GITHUB_TOKEN"))

	viper.SetDefault("zappr.url", os.Getenv("ZAPPR_URL"))
	viper.SetDefault("zappr.usezapprgithubcredentials", true)
	viper.BindEnv("ZAPPR_USE_APP_CREDENTIALS", "zappr.usezapprgithubcredentials")

	err := viper.ReadInConfig()
	if err != nil {
		return ctx, fmt.Errorf("could not read configurations: %w", err)
	}

	config := Spec{}
	err = viper.Unmarshal(&config)
	if err != nil {
		return ctx, fmt.Errorf("could not unmarshal config file: %w", err)
	}

	return context.WithValue(ctx, configKey, &config), nil
}

// WithContext returns a logrus logger from the context
func WithContext(ctx context.Context) *Spec {
	if ctx == nil {
		return nil
	}

	if ctxConfig, ok := ctx.Value(configKey).(*Spec); ok {
		return ctxConfig
	}

	return nil
}

// OverrideConfig writes a new context with the the new configuration
func OverrideConfig(ctx context.Context, config *Spec) context.Context {
	return context.WithValue(ctx, configKey, config)
}
