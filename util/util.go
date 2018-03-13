package util

import (
	"os"

	"github.com/sirupsen/logrus"
)

func SetupLogging(log *logrus.Logger, logJSON bool, logLevel string) {
	if logJSON {
		log.Formatter = new(logrus.JSONFormatter)
	}

	log.Out = os.Stdout

	switch logLevel {
	case "debug":
		log.Level = logrus.DebugLevel
	case "info":
		log.Level = logrus.InfoLevel
	case "warn":
		log.Level = logrus.WarnLevel
	default:
		log.Level = logrus.ErrorLevel
	}
}
