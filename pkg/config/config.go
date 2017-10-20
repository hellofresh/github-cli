package config

import (
	"github.com/spf13/viper"
)

type (
	// Spec represents the global app configuration
	Spec struct {
		PublicKeyPath string
		Github        Github
		PullApprove   PullApprove
		LogLevel      string
	}

	// PullApprove represents teh Pull Approve configurations
	PullApprove struct {
		Token               string
		Filename            string
		ProtectedBranchName string
	}

	// Github represents the github configurations
	Github struct {
		Organization  string
		Token         string
		Teams         []*Team
		Collaborators []*Collaborator
		Labels        []*Label
		Webhooks      []*Webhook
		Protections   map[string][]string
		// RemoveDefaultLabels Remove GitHub's default labels?
		RemoveDefaultLabels bool
	}

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

// Load loads all the configurations
func Load(cfgFile string) (*Spec, error) {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".github")
		viper.AddConfigPath("/etc/github")
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config *Spec
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return config, nil
}
