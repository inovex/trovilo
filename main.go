package main

import (
	"context"

	"github.com/alecthomas/kingpin"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	"github.com/sirupsen/logrus"
)

var (
	configFile     = kingpin.Flag("config", "YAML configuration file.").Required().ExistingFile()
	kubeConfigFile = kingpin.Flag("kubeconfig", "Optional kubectl configuration file. If undefined we expect trovilo to be running in a pod.").ExistingFile()
	logLevel       = kingpin.Flag("log-level", "Specify log level (debug, info, warn, error).").Default("info").String()
	logJSON        = kingpin.Flag("log-json", "Enable JSON-formatted logging on STDOUT.").Bool()
	log            = logrus.New()
)

func main() {
	// Prepare cmd line parser
	kingpin.CommandLine.Help = "Trovilo collects and prepares files from Kubernetes ConfigMaps for Prometheus & friends"
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	// Prepare logging
	setupLogging()

	// Prepare app configuration
	config, err := getConfig(*configFile)
	if err != nil {
		log.Fatalf("Error parsing config file: %s", err)
	}
	log.WithFields(logrus.Fields{"config": config}).Debug("Successfully loaded configuration")

	// Prepare Kubernetes client
	client, err := getClient(kubeConfigFile)
	if err != nil {
		log.WithError(err).Fatal("Failed to load Kubernetes client")
	}
	log.Debug("Successfully loaded Kubernetes client")

	log.Debug("Testing Kubernetes connectivity by fetching list of nodes..")
	if err := client.List(context.Background(), "", new(corev1.NodeList)); err != nil {
		log.WithError(err).Fatal("Failed to test Kubernetes connectivity")
	}
	log.Debug("Successfully tested Kubernetes connectivity")

	// Setup ConfigMap watcher
	log.WithFields(logrus.Fields{
		"namespace": config.Namespace,
	}).Info("Starting to watch for new/modified/deleted ConfigMaps (empty namespace string means all accessible namespaces)")

	watcher, err := client.Watch(context.Background(), config.Namespace, new(corev1.ConfigMap))
	if err != nil {
		log.WithError(err).Fatal("Failed to load Kubernetes ConfigMap watcher")
	}
	defer watcher.Close()

	for {
		cm := new(corev1.ConfigMap)
		eventType, err := watcher.Next(cm)
		if err != nil {
			log.WithError(err).Fatal("Kubernetes ConfigMap watcher encountered an error. Exit..") //TODO is it necessary to exit?
		}

		for jobPos := range config.Jobs {
			// Check whether ConfigMap matches to our expected labels
			if !compareCMLabels(&config.Jobs[jobPos].Labels, &cm.Metadata.Labels) {
				// It doesn't match. Make sure it doesn't exist (anymore) in filesystem
				if isCMAlreadyRegistered(cm, &config.Jobs[jobPos].TargetDir) {
					log.WithFields(logrus.Fields{
						"job":       config.Jobs[jobPos].Name,
						"configmap": *cm.Metadata.Name,
						"namespace": *cm.Metadata.Namespace,
					}).Info("ConfigMap has been deleted from namepace, thus removing in target directory too")

					removedFiles, err := removeCMfromTargetDir(cm, &config.Jobs[jobPos].TargetDir)

					if err == nil {
						log.WithFields(logrus.Fields{
							"job":          config.Jobs[jobPos].Name,
							"configmap":    *cm.Metadata.Name,
							"namespace":    *cm.Metadata.Namespace,
							"removedFiles": removedFiles,
						}).Info("Successfully deleted ConfigMap from namepace")
					} else {
						log.WithFields(logrus.Fields{
							"job":          config.Jobs[jobPos].Name,
							"configmap":    *cm.Metadata.Name,
							"namespace":    *cm.Metadata.Namespace,
							"removedFiles": removedFiles,
							"error":        err,
						}).Fatal("Failed to delete ConfigMap from namepace")
					}

					continue
				}

				// Leave some debug information and just switch to the next ConfigMap
				log.WithFields(logrus.Fields{
					"job":            config.Jobs[jobPos].Name,
					"configmap":      *cm.Metadata.Name,
					"namespace":      *cm.Metadata.Namespace,
					"actualLabels":   cm.Metadata.Labels,
					"expectedLabels": config.Jobs[jobPos].Labels,
					"eventType":      eventType,
				}).Debug("ConfigMap did not contain expected labels, skipping..")
				continue
			}

			log.WithFields(logrus.Fields{
				"job":            config.Jobs[jobPos].Name,
				"configmap":      *cm.Metadata.Name,
				"namespace":      *cm.Metadata.Namespace,
				"actualLabels":   cm.Metadata.Labels,
				"expectedLabels": config.Jobs[jobPos].Labels,
				"eventType":      eventType,
			}).Info("Found matching ConfigMap")

			if len(config.Jobs[jobPos].Verify) > 0 {
				log.WithFields(logrus.Fields{
					"job":         config.Jobs[jobPos].Name,
					"configmap":   *cm.Metadata.Name,
					"namespace":   *cm.Metadata.Namespace,
					"verifySteps": config.Jobs[jobPos].Verify,
				}).Debug("Verifying ConfigMap against validity")

				// Verify validity of ConfigMap files
				verifiedFiles, latestOutput, err := verifyCM(cm, &config.Jobs[jobPos].Verify)
				if err != nil {
					log.WithFields(logrus.Fields{
						"job":           config.Jobs[jobPos].Name,
						"configmap":     *cm.Metadata.Name,
						"namespace":     *cm.Metadata.Namespace,
						"verifySteps":   config.Jobs[jobPos].Verify,
						"verifiedFiles": verifiedFiles,
						"latestOutput":  latestOutput,
						"error":         err,
					}).Warn("Failed to verify ConfigMap, there's something wrong with it, so we skip it..") //TODO document that we won't remove files that aren't valid anymore
					continue
				} else {
					log.WithFields(logrus.Fields{
						"job":           config.Jobs[jobPos].Name,
						"configmap":     *cm.Metadata.Name,
						"namespace":     *cm.Metadata.Namespace,
						"verifySteps":   config.Jobs[jobPos].Verify,
						"verifiedFiles": verifiedFiles,
					}).Debug("Successfully verified ConfigMap, ready to register")
				}
			}

			// ConfigMap has been verified, write files it to filesystem
			registeredFiles, err := registerCM(cm, &config.Jobs[jobPos].TargetDir)
			if err == nil {
				log.WithFields(logrus.Fields{
					"job":             config.Jobs[jobPos].Name,
					"configmap":       *cm.Metadata.Name,
					"namespace":       *cm.Metadata.Namespace,
					"eventType":       eventType,
					"registeredFiles": registeredFiles,
				}).Info("Successfully registered ConfigMap")
			} else {
				log.WithFields(logrus.Fields{
					"job":             config.Jobs[jobPos].Name,
					"configmap":       *cm.Metadata.Name,
					"namespace":       *cm.Metadata.Namespace,
					"eventType":       eventType,
					"registeredFiles": registeredFiles,
					"error":           err,
				}).Fatal("Failed to register ConfigMap")
			}
		}
	}
}
