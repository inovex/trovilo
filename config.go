package main

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type verifyStep struct {
	Name string   `yaml:"name"`
	Cmd  []string `yaml:"cmd"`
}
type jobConfig struct {
	Name string `yaml:"name"`

	Labels    map[string]string `yaml:"labels"`
	Verify    []verifyStep      `yaml:"verify"`
	TargetDir string            `yaml:"target-dir"`
}

type config struct {
	Namespace string      `yaml:"namespace"`
	Jobs      []jobConfig `yaml:"jobs"`
}

func getConfig(configFile string) (config, error) {
	yamlFile, err := ioutil.ReadFile(configFile)

	if yamlFile == nil {
		log.WithError(err).Fatalf("Error reading config file")
	}

	config := config{}
	err = yaml.Unmarshal(yamlFile, &config)

	return config, err
}
