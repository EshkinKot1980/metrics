package file

import (
	"fmt"
	"os"
	"path/filepath"
)

func fileBackup(fileName string) {
	if fileExists(fileName) {
		os.Rename(fileName, fileName+".backup")
	}
}

func fileRestore(fileName string) {
	if fileExists(fileName + ".backup") {
		os.Rename(fileName+".backup", fileName)
	}
}

func removeBackup(fileName string) {
	if fileExists(fileName + ".backup") {
		os.Remove(fileName + ".backup")
	}
}

func fileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func checkFileDir(fileName string) error {
	dirPath := filepath.Dir(fileName)

	info, err := os.Stat(dirPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if info.IsDir() {
			return nil
		} else {
			return fmt.Errorf("file %s already exists and it is not a directory", dirPath)
		}
	}

	return os.MkdirAll(dirPath, 0775)
}
