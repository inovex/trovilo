package config

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"

	yaml "gopkg.in/yaml.v2"
)

type VerifyStepCmd []string

type VerifyStep struct {
	Name string        `yaml:"name"`
	Cmd  VerifyStepCmd `yaml:"cmd"`
}

type PostDeployActionCmd []string

type PostDeployAction struct {
	Name string              `yaml:"name"`
	Cmd  PostDeployActionCmd `yaml:"cmd"`
}

type JobConfig struct {
	Name string `yaml:"name"`

	Selector   map[string]string  `yaml:"selector"`
	Verify     []VerifyStep       `yaml:"verify"`
	TargetDir  string             `yaml:"target-dir"`
	Flatten    bool               `yaml:"flatten"`
	PostDeploy []PostDeployAction `yaml:"post-deploy"`
}

type Config struct {
	Namespace string      `yaml:"namespace"`
	Jobs      []JobConfig `yaml:"jobs"`
}

// GetConfig Translates the YAML main configuration file into Config struct
func GetConfig(log *logrus.Logger, configFile string) (Config, error) {
	yamlFile, err := ioutil.ReadFile(configFile)

	if yamlFile == nil {
		log.WithError(err).Fatalf("Error reading config file")
	}

	config := Config{}
	err = yaml.Unmarshal(yamlFile, &config)

	return config, err
}
