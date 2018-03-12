package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

func setupLogging() {
	if *logJSON {
		log.Formatter = new(logrus.JSONFormatter)
	}

	log.Out = os.Stdout

	if *logLevel == "debug" {
		log.Level = logrus.DebugLevel
	} else if *logLevel == "info" {
		log.Level = logrus.InfoLevel
	} else if *logLevel == "warn" {
		log.Level = logrus.WarnLevel
	} else {
		log.Level = logrus.ErrorLevel
	}
}
