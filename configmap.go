package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	corev1 "github.com/ericchiang/k8s/apis/core/v1"
)

func writeFile(file *string, contents *string) error {
	os.MkdirAll(filepath.Dir(*file), 0755)
	return ioutil.WriteFile(*file, []byte(*contents), 0644)
}

func writeOSFile(file *os.File, contents *string) (int, error) {
	os.MkdirAll(filepath.Dir(file.Name()), 0755)
	return file.Write([]byte(*contents))
}

func verifyCM(configMap *corev1.ConfigMap, verifySteps *[]verifyStep) (map[string]string, string, error) {
	verifiedFiles := map[string]string{}

	for file, fileContents := range configMap.Data {
		for step := range *verifySteps {
			if len((*verifySteps)[step].Cmd) != 0 {
				var args []string

				// Prepare (temporary) file to verify
				tempFile, err := ioutil.TempFile("", fmt.Sprintf("trovilo-%s-", file))
				if err != nil {
					return verifiedFiles, "", err
				}
				_, err = writeOSFile(tempFile, &fileContents)
				if err != nil {
					return verifiedFiles, "", err
				}

				for cmdPos := range (*verifySteps)[step].Cmd {
					arg := (*verifySteps)[step].Cmd[cmdPos]
					if strings.Contains(arg, "%s") {
						args = append(args, fmt.Sprintf(arg, tempFile.Name()))
					} else {
						args = append(args, arg)
					}
				}
				cmd := exec.Command(args[0], args[1:]...)
				out, err := cmd.CombinedOutput()
				niceOutput := strings.TrimSpace(string(out))

				// Remove the file, regardless of the verification result
				os.Remove(tempFile.Name())

				if err != nil {
					// Immediately abort if there's just one piece of the configmap that is invalid
					return verifiedFiles, niceOutput, err
				}

				verifiedFiles[file] = niceOutput
			}
		}

	}

	return verifiedFiles, "", nil
}

func registerCM(configMap *corev1.ConfigMap, targetDir *string) ([]string, error) {
	var registeredFiles []string

	for file, fileContents := range configMap.Data {
		targetFile := filepath.Join(*targetDir, *configMap.Metadata.Namespace, *configMap.Metadata.Name, file)
		registeredFiles = append(registeredFiles, targetFile)

		err := writeFile(&targetFile, &fileContents)
		if err != nil {
			return registeredFiles, err
		}
	}

	return registeredFiles, nil
}

// Helper function that checks whether we already know this ConfigMap
func isCMAlreadyRegistered(configMap *corev1.ConfigMap, targetDir *string) bool {
	for file := range configMap.Data {
		targetFile := filepath.Join(*targetDir, *configMap.Metadata.Namespace, *configMap.Metadata.Name, file)

		_, err := os.Stat(targetFile)
		if err == nil {
			return true
		}
	}
	return false
}

func removeCMfromTargetDir(configMap *corev1.ConfigMap, targetDir *string) ([]string, error) {
	var removedFiles []string

	for file := range configMap.Data {
		targetFile := filepath.Join(*targetDir, file)
		removedFiles = append(removedFiles, *configMap.Metadata.Namespace, *configMap.Metadata.Name, targetFile)

		err := os.Remove(targetFile)
		if err != nil {
			return removedFiles, err
		}
	}

	return removedFiles, nil
}
