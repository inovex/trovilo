package logging

import (
	"io/ioutil"
	"os"
	"path/filepath"

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

func WriteFile(file string, contents []byte) error {
	err := os.MkdirAll(filepath.Dir(file), 0755)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(file, contents, 0644)
}

func WriteOSFile(file *os.File, contents []byte) error {
	err := os.MkdirAll(filepath.Dir(file.Name()), 0755)
	if err != nil {
		return err
	}

	_, err = file.Write(contents)
	return err
}

func DeleteFile(file string) error {
	return os.Remove(file)
}
