package file

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/EshkinKot1980/metrics/internal/server"
	"github.com/EshkinKot1980/metrics/internal/server/storage"
)

type config = server.FileStorageConfig

type Logger interface {
	Error(message string, err error)
}

type FileStorage struct {
	config   config
	logger   Logger
	halt     chan struct{}
	cmx      sync.RWMutex
	counters map[string]int64
	gauges   map[string]float64
}

func New(c config, l Logger) (*FileStorage, error) {
	s := &FileStorage{
		config:   c,
		logger:   l,
		halt:     make(chan struct{}),
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}

	return s, s.start()
}

func (s *FileStorage) PutCounter(name string, increment int64) int64 {
	s.cmx.Lock()
	defer func() {
		s.sync()
		s.cmx.Unlock()
	}()

	s.counters[name] += increment
	return s.counters[name]
}

func (s *FileStorage) PutGauge(name string, value float64) {
	s.cmx.Lock()
	defer func() {
		s.sync()
		s.cmx.Unlock()
	}()

	s.gauges[name] = value
}

func (s *FileStorage) GetCounter(name string) (int64, error) {
	s.cmx.RLock()
	defer s.cmx.RUnlock()

	v, ok := s.counters[name]
	if !ok {
		return v, storage.ErrCounterNotFound
	}

	return v, nil
}

func (s *FileStorage) GetGauge(name string) (float64, error) {
	s.cmx.RLock()
	defer s.cmx.RUnlock()

	v, ok := s.gauges[name]
	if !ok {
		return v, storage.ErrGaugeNotFound
	}

	return v, nil
}

func (s *FileStorage) Halt() {
	close(s.halt)
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

	s.intervalSync()
	return nil
}

func (s *FileStorage) sync() {
	if s.config.Interval == 0 {
		s.flush()
	}
}

func (s *FileStorage) intervalSync() {
	if s.config.Interval == 0 {
		return
	}

	go func() {
		interval := time.Duration(s.config.Interval) * time.Second
		cancel := false
		for {
			select {
			case <-time.After(interval):
			case <-s.halt:
				cancel = true
			}

			s.cmx.RLock()
			s.flush()
			s.cmx.RUnlock()

			if cancel {
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
