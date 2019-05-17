package main

import (
	"context"

	"github.com/alecthomas/kingpin"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	"github.com/inovex/trovilo/client"
	"github.com/inovex/trovilo/config"
	"github.com/inovex/trovilo/configmap"
	"github.com/inovex/trovilo/logging"
	"github.com/sirupsen/logrus"
)

var build string

func processPostDeployActions(logEntryBase *logrus.Entry, postDeployActions []config.PostDeployAction) {
	for _, postDeployAction := range postDeployActions {
		output, err := configmap.RunPostDeployActionCmd(postDeployAction.Cmd)
		logEntry := *logEntryBase.WithFields(logrus.Fields{
			"postDeployAction": postDeployAction,
			"output":           output,
		})
		if err != nil {
			logEntry.WithError(err).Error("Failed to executed postDeployAction command")
		} else {
			logEntry.Info("Successfully executed postDeployAction command")
		}
	}
}

func main() {
	// Prepare cmd line parser
	var configFile = kingpin.Flag("config", "YAML configuration file.").Required().ExistingFile()
	var kubeConfigFile = kingpin.Flag("kubeconfig", "Optional kubectl configuration file. If undefined we expect trovilo is running in a pod.").ExistingFile()
	var logLevel = kingpin.Flag("log-level", "Specify log level (debug, info, warn, error).").Default("info").String()
	var logJSON = kingpin.Flag("log-json", "Enable JSON-formatted logging on STDOUT.").Bool()

	kingpin.CommandLine.Help = "Trovilo collects and prepares files from Kubernetes ConfigMaps for Prometheus & friends"
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Version(build)
	kingpin.CommandLine.VersionFlag.Short('v')
	kingpin.Parse()

	// Prepare logging
	log := logrus.New()
	logging.SetupLogging(log, *logJSON, *logLevel)

	// Prepare app configuration
	config, err := config.GetConfig(log, *configFile)
	if err != nil {
		log.WithError(err).Fatal("Error parsing config file")
	}
	log.WithFields(logrus.Fields{"config": config}).Debug("Successfully loaded configuration")

	// Prepare Kubernetes client
	client, err := client.GetClient(*kubeConfigFile)
	if err != nil {
		log.WithError(err).Fatal("Failed to load Kubernetes client")
	}
	log.Debug("Successfully loaded Kubernetes client")

	log.Debug("Testing Kubernetes connectivity by fetching list of configmaps..")
	if err := client.List(context.Background(), config.Namespace, new(corev1.ConfigMapList)); err != nil {
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

		for _, job := range config.Jobs {
			logEntryBase := log.WithFields(logrus.Fields{
				"job":       job.Name,
				"configmap": *cm.Metadata.Name,
				"namespace": *cm.Metadata.Namespace,
			})
			logEntryWithSelectors := logEntryBase.WithFields(logrus.Fields{
				"actualLabels":   cm.Metadata.Labels,
				"expectedLabels": job.Selector,
				"eventType":      eventType,
			})
			// Check whether ConfigMap matches to our expected labels
			match := configmap.CompareCMLabels(job.Selector, cm.Metadata.Labels)
			if !match || eventType == "DELETED" {
				// It doesn't match. Make sure it doesn't exist in the filesystem (anymore)
				if configmap.IsCMAlreadyRegistered(cm, job.TargetDir, job.Flatten) {
					logEntryBase.Info("ConfigMap has been deleted from namepace, thus removing in target directory too")
					removedFiles, err := configmap.RemoveCMfromTargetDir(cm, job.TargetDir, job.Flatten)

					logEntry := logEntryBase.WithField("removedFiles", removedFiles)
					if err != nil {
						logEntry.WithError(err).Fatal("Failed to delete ConfigMap from namepace")
					} else {
						logEntry.Info("Successfully deleted ConfigMap from namepace")
					}

					// Deleting the configmap also triggers post-deploy actions
					if len(job.PostDeploy) > 0 {
						processPostDeployActions(logEntryBase, job.PostDeploy)
					}
					continue
				}

				// Leave some debug information and just switch to the next ConfigMap
				logEntryWithSelectors.Debug("ConfigMap did not contain expected labels, skipping..")
				continue
			}

			logEntryBase.WithFields(logrus.Fields{
				"actualLabels":   cm.Metadata.Labels,
				"expectedLabels": job.Selector,
				"eventType":      eventType,
			}).Info("Found matching ConfigMap")

			if len(job.Verify) > 0 {
				logEntryBase.WithField("verifySteps", job.Verify).Debug("Verifying ConfigMap against validity")

				// Verify validity of ConfigMap files
				verifiedFiles, latestOutput, err := configmap.VerifyCM(cm, job.Verify)
				if err != nil {
					logEntryBase.WithFields(logrus.Fields{
						"verifySteps":   job.Verify,
						"verifiedFiles": verifiedFiles,
						"latestOutput":  latestOutput,
						"error":         err,
					}).Warn("Failed to verify ConfigMap, there's something wrong with it, so we skip it..") //TODO document that we won't remove files that aren't valid anymore
					continue
				} else {
					logEntryBase.WithFields(logrus.Fields{
						"verifySteps":   job.Verify,
						"verifiedFiles": verifiedFiles,
					}).Debug("Successfully verified ConfigMap, ready to register")
				}
			}

			// ConfigMap has been verified, write files to filesystem
			registeredFiles, err := configmap.RegisterCM(cm, job.TargetDir, job.Flatten)
			logEntry := logEntryBase.WithFields(logrus.Fields{
				"eventType":       eventType,
				"registeredFiles": registeredFiles,
			})
			if err != nil {
				logEntry.WithError(err).Fatal("Failed to register ConfigMap")
			} else {
				logEntry.Info("Successfully registered ConfigMap")
			}

			// ConfigMap has ben registered, now run (optional) user-defined post deployment actions
			if len(job.PostDeploy) > 0 {
				processPostDeployActions(logEntryBase, job.PostDeploy)
			}
		}
	}
}
