package storage

import (
	"sync"
	"github.com/EshkinKot1980/metrics/internal/agent/model"
)

type MemoryStorage struct {
	mx 	 sync.Mutex
	counters map[string]int64
	gauges   map[string]float64 
}

func New() *MemoryStorage {
	return &MemoryStorage{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (s *MemoryStorage) PutLock() {
	s.mx.Lock()
}

func (s *MemoryStorage) PutUnlock() {
	s.mx.Unlock()
}

func (s *MemoryStorage) PutCounter(c model.Counter) {
	s.counters[c.Name] += c.Value
}

func (s *MemoryStorage) PutGauge(g model.Gauge) {
	s.gauges[g.Name] = g.Value
}

func (s *MemoryStorage) Pull() ([]model.Counter, []model.Gauge) {
	s.mx.Lock()
	defer func() {
		s.counters = make(map[string]int64)
		s.gauges = make(map[string]float64)
		s.mx.Unlock()
	}()

	counters := make([]model.Counter, len(s.counters))
	i := 0 
	for n, v := range s.counters {
		counters[i] = model.Counter{Name: n, Value: v}
		i++
	}

	gauges := make([]model.Gauge, len(s.gauges))
	i = 0 
	for n, v := range s.gauges {		
		gauges[i] = model.Gauge{Name: n, Value: v}
		i++
	}
	
	return counters, gauges
}