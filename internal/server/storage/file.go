package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/EshkinKot1980/metrics/internal/server/config"
)

type FSconfig = config.FileStorageConfig

type Logger interface {
	Error(message string, err error)
}

type FileStorage struct {
	MemoryStorage
	config FSconfig
	logger Logger
	haltCh chan struct{}
	syncCh chan struct{}
}

func NewFileStorage(c FSconfig, l Logger) (*FileStorage, error) {
	s := &FileStorage{
		config: c,
		logger: l,
		haltCh: make(chan struct{}),
		syncCh: make(chan struct{}),
	}

	s.counters = make(map[string]int64)
	s.gauges = make(map[string]float64)

	return s, s.start()
}

func (s *FileStorage) PutCounter(c Counter) (Counter, error) {
	defer s.sync()
	return s.MemoryStorage.PutCounter(c)
}

func (s *FileStorage) PutGauge(g Gauge) error {
	defer s.sync()
	return s.MemoryStorage.PutGauge(g)
}

func (s *FileStorage) PutMetrics(ctx context.Context, counters []Counter, gauges []Gauge) error {
	defer s.sync()
	return s.MemoryStorage.PutMetrics(ctx, counters, gauges)
}

func (s *FileStorage) Halt() {
	close(s.haltCh)
	<-time.After(time.Duration(1) * time.Second)
}

func (s *FileStorage) start() error {
	if err := checkFileDir(s.config.Path); err != nil {
		return err
	}

	if s.config.Restore {
		if err := s.load(); err != nil {
			return err
		}
	}

	s.startSync()
	return nil
}

func (s *FileStorage) sync() {
	select {
	case s.syncCh <- struct{}{}:
	default:
	}
}

func (s *FileStorage) startSync() {
	if s.config.Interval == 0 {
		go func() {
			for {
				<-s.syncCh
				s.flush()
			}
		}()

		return
	}

	go func() {
		defer s.flush()
		interval := time.Duration(s.config.Interval) * time.Second
		for {
			select {
			case <-time.After(interval):
				s.flush()
			case <-s.haltCh:
				return
			}
		}
	}()
}

type fileData struct {
	Counters map[string]int64
	Gauges   map[string]float64
}

func (s *FileStorage) flush() {
	s.cmx.RLock()
	defer s.cmx.RUnlock()

	fileName := s.config.Path
	undo := false
	data := fileData{
		Counters: s.counters,
		Gauges:   s.gauges,
	}

	fileBackup(fileName)
	file, err := os.Create(fileName)
	if err != nil {
		s.logger.Error("failed to open file", err)
		fileRestore(fileName)
		return
	}
	defer func() {
		file.Close()
		if undo {
			fileRestore(fileName)
		}
	}()

	if err := json.NewEncoder(file).Encode(data); err != nil {
		undo = true
		s.logger.Error("failed to write file", err)
		return
	}
	removeBackup(fileName)
}

func (s *FileStorage) load() error {
	if !fileExists(s.config.Path) {
		return nil
	}

	file, err := os.Open(s.config.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	var data fileData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	s.gauges = data.Gauges
	s.counters = data.Counters
	return nil
}

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
