// Модуль storage реализует хранение данных на сервере.
package storage

import (
	"context"
	"sync"
)

// Хранилище метрик в пямяти, реализует интерфейс Storage, используется в тестах.
type MemoryStorage struct {
	cmx      sync.RWMutex
	counters map[string]int64
	gauges   map[string]float64
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (s *MemoryStorage) PutCounter(c Counter) (Counter, error) {
	s.cmx.Lock()
	defer s.cmx.Unlock()
	s.counters[c.Name] += c.Value
	c.Value = s.counters[c.Name]
	return c, nil
}

func (s *MemoryStorage) PutGauge(g Gauge) error {
	s.cmx.Lock()
	defer s.cmx.Unlock()
	s.gauges[g.Name] = g.Value
	return nil
}

func (s *MemoryStorage) GetCounter(c Counter) (Counter, error) {
	s.cmx.RLock()
	defer s.cmx.RUnlock()

	v, ok := s.counters[c.Name]
	if !ok {
		return c, ErrCounterNotFound
	}

	c.Value = v
	return c, nil
}

func (s *MemoryStorage) GetGauge(g Gauge) (Gauge, error) {
	s.cmx.RLock()
	defer s.cmx.RUnlock()

	v, ok := s.gauges[g.Name]
	if !ok {
		return g, ErrGaugeNotFound
	}

	g.Value = v
	return g, nil
}

func (s *MemoryStorage) PutMetrics(ctx context.Context, counters []Counter, gauges []Gauge) error {
	s.cmx.Lock()
	defer s.cmx.Unlock()

	for _, c := range counters {
		s.counters[c.Name] += c.Value
	}

	for _, g := range gauges {
		s.gauges[g.Name] = g.Value
	}

	return nil
}

func (s *MemoryStorage) Ping() bool {
	return false
}

func (s *MemoryStorage) Halt() {}
