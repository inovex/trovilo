package filesystem

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

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
