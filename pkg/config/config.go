package config

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/errwrap"
	"github.com/hellofresh/github-cli/pkg/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type (
	configKeyType int

	// Spec represents the global app configuration
	Spec struct {
		Github        Github
		GithubTestOrg Github
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

	err := viper.ReadInConfig()
	if err != nil {
		return ctx, errwrap.Wrapf("could not read configurations: {{err}}", err)
	}

	config := Spec{}
	err = viper.Unmarshal(&config)
	if err != nil {
		return ctx, errwrap.Wrapf("could not unmarshal config file: {{err}}", err)
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
