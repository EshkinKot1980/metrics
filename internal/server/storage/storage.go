package storage

import (
	"context"
	"errors"

	"github.com/EshkinKot1980/metrics/internal/common/models"
)

var (
	ErrCounterNotFound = errors.New("counter metric not found")
	ErrGaugeNotFound   = errors.New("gauge metric not found")
)

type Storage interface {
	Halt()
	PutCounter(name string, increment int64) (int64, error)
	PutGauge(name string, value float64) error
	PutMetrics(ctx context.Context, metrics []models.Metrics) error
	GetCounter(name string) (int64, error)
	GetGauge(name string) (float64, error)
}
