package storage

import (
	"context"
	"errors"

	"github.com/EshkinKot1980/metrics/internal/server/config"
)

var (
	ErrCounterNotFound = errors.New("counter metric not found")
	ErrGaugeNotFound   = errors.New("gauge metric not found")
)

type Gauge struct {
	Name  string
	Value float64
}

type Counter struct {
	Name  string
	Value int64
}

type Storage interface {
	GetCounter(c Counter) (Counter, error)
	GetGauge(g Gauge) (Gauge, error)
	PutCounter(c Counter) (Counter, error)
	PutGauge(g Gauge) error
	PutMetrics(ctx context.Context, counters []Counter, gauges []Gauge) error
	Halt()
	Ping() bool
}

func New(cfg *config.Config, l Logger) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		return NewDBStorage(cfg.DatabaseDSN)
	}

	return NewFileStorage(cfg.FileCfg, l)
}
