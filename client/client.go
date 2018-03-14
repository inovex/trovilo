package client

import (
	"io/ioutil"

	"github.com/ericchiang/k8s"
	yaml "gopkg.in/yaml.v2"
)

// loadKubernetesConfig parses a kubeconfig from a file and returns a Kubernetes
// client. It does not support extensions or client auth providers.
func loadKubernetesConfig(kubeConfigFile string) (*k8s.Client, error) {
	var data []byte
	var err error

	if data, err = ioutil.ReadFile(kubeConfigFile); err != nil {
		return nil, err
	}

	// Unmarshal YAML into a Kubernetes config object.
	var config k8s.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return k8s.NewClient(&config)
}

// GetClient returns a k8s client to interact with k8s
func GetClient(kubeConfigFile string) (*k8s.Client, error) {
	if kubeConfigFile != "" {
		return loadKubernetesConfig(kubeConfigFile)
	}

	return k8s.NewInClusterClient()
}
