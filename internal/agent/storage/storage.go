package storage

import (
	"github.com/EshkinKot1980/metrics/internal/agent/model"
	"sync"
)

type MemoryStorage struct {
	mx       sync.Mutex
	counters map[string]int64
	gauges   map[string]float64
}

func New() *MemoryStorage {
	return &MemoryStorage{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (s *MemoryStorage) Put(c []model.Counter, g []model.Gauge) {
	s.mx.Lock()
	defer s.mx.Unlock()

	for _, m := range c {
		s.counters[m.Name] += m.Value
	}

	for _, m := range g {
		s.gauges[m.Name] = m.Value
	}
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
