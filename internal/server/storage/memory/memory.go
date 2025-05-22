package memory

import (
	"github.com/EshkinKot1980/metrics/internal/server/storage"
	"sync"
)

type MemoryStorage struct {
	cmx      sync.RWMutex
	counters map[string]int64
	gauges   map[string]float64
}

func New() *MemoryStorage {
	return &MemoryStorage{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (s *MemoryStorage) PutCounter(name string, increment int64) int64 {
	s.cmx.Lock()
	defer s.cmx.Unlock()
	s.counters[name] += increment
	return s.counters[name]
}

func (s *MemoryStorage) PutGauge(name string, value float64) {
	s.gauges[name] = value
}

func (s *MemoryStorage) GetCounter(name string) (int64, error) {
	s.cmx.RLock()
	defer s.cmx.RUnlock()

	v, ok := s.counters[name]
	if !ok {
		return v, storage.ErrCounterNotFound
	}

	return v, nil
}

func (s *MemoryStorage) GetGauge(name string) (float64, error) {
	v, ok := s.gauges[name]
	if !ok {
		return v, storage.ErrGaugeNotFound
	}

	return v, nil
}
