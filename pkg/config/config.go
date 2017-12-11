package config

import (
	"os"
	"os/user"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type (
	// Spec represents the global app configuration
	Spec struct {
		Github        Github
		GithubTestOrg Github
		PullApprove   PullApprove
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

// Load loads all the configurations
func Load(cfgFile string) (*Spec, error) {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".github")
		viper.AddConfigPath(".")
		viper.AddConfigPath(homeDir())
	}

	viper.SetDefault("github.token", os.Getenv("GITHUB_TOKEN"))
	viper.SetDefault("githubtestorg.token", os.Getenv("GITHUB_TOKEN"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config *Spec
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	return usr.HomeDir
}
