package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type FileAuditor struct {
	path  string
	loger Logger
	mx    *sync.Mutex
}

func NewFileAuditor(filePath string, l Logger) (*FileAuditor, error) {
	if err := checkFile(filePath); err != nil {
		return nil, err
	}

	return &FileAuditor{path: filePath, loger: l, mx: &sync.Mutex{}}, nil
}

func (a *FileAuditor) Stop() {
	a.mx.Lock()
}

func (a *FileAuditor) Handle(e Event) {
	a.mx.Lock()
	defer a.mx.Unlock()

	file, err := os.OpenFile(a.path, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		a.loger.Error("failed to open file", err)
		return
	}
	defer file.Close()

	data, err := json.Marshal(e)
	if err != nil {
		a.loger.Error("failed to json encode data", err)
		return
	}
	eol := []byte("\n")
	data = append(data, eol...)

	_, err = file.Write(data)
	if err != nil {
		a.loger.Error("failed to write data to file", err)
		return
	}
}

func checkFile(fileName string) error {
	info, err := os.Stat(fileName)

	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("invalid file path, %w", err)
		}

		if err := checkFileDir(fileName); err != nil {
			return fmt.Errorf("invalid file path, %w", err)
		}
	} else {
		if info.IsDir() {
			return fmt.Errorf("file %s already exists and it is a directory", fileName)
		}
	}

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create or open file, %w", err)
	}
	file.Close()

	return nil
}

func checkFileDir(fileName string) error {
	dirPath := filepath.Dir(fileName)

	info, err := os.Stat(dirPath)
	if err == nil && info.IsDir() {
		return nil
	}

	return os.MkdirAll(dirPath, 0775)
}
