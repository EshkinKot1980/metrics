package service

import (
	"context"
	"errors"
	"time"

	"github.com/EshkinKot1980/metrics/internal/common/models"
	"github.com/EshkinKot1980/metrics/internal/server/audit"
	"github.com/EshkinKot1980/metrics/internal/server/storage"
)

type ContextKey string

const KeyClientIP ContextKey = "clientIP"

var ErrMetricNotFound = errors.New("metric not found")

type Storage interface {
	GetCounter(c storage.Counter) (storage.Counter, error)
	GetGauge(g storage.Gauge) (storage.Gauge, error)
	PutCounter(c storage.Counter) (storage.Counter, error)
	PutGauge(g storage.Gauge) error
	PutMetrics(ctx context.Context, counters []storage.Counter, gauges []storage.Gauge) error
}

type Logger interface {
	Error(message string, err error)
}

type Auditor interface {
	Rise(e audit.Event)
}

type MetricService struct {
	storage Storage
	logger  Logger
	auditor Auditor
}

func NewMetricService(s Storage, l Logger, a Auditor) *MetricService {
	return &MetricService{storage: s, logger: l, auditor: a}
}

func (s *MetricService) Fill(metric models.Metrics) (models.Metrics, error) {
	var (
		counter storage.Counter
		gauge   storage.Gauge
		err     error
	)

	switch metric.MType {
	case models.TypeGauge:
		gauge, err = s.storage.GetGauge(storage.Gauge{Name: metric.ID})
		metric.Value = &gauge.Value
		metric.Delta = nil
	case models.TypeCounter:
		counter, err = s.storage.GetCounter(storage.Counter{Name: metric.ID})
		metric.Delta = &counter.Value
		metric.Value = nil
	default:
		err = models.ErrInvalidMetricType
	}

	if err != nil {
		switch err {
		case storage.ErrCounterNotFound, storage.ErrGaugeNotFound:
			err = ErrMetricNotFound
		default:
			s.logger.Error("failed to get metric", err)
		}
	}

	return metric, err
}

func (s *MetricService) Put(metric models.Metrics) (models.Metrics, error) {
	var err error

	switch metric.MType {
	case models.TypeGauge:
		gauge := storage.Gauge{Name: metric.ID, Value: *metric.Value}
		err = s.storage.PutGauge(gauge)
		metric.Delta = nil
	case models.TypeCounter:
		counter := storage.Counter{Name: metric.ID, Value: *metric.Delta}
		counter, err = s.storage.PutCounter(counter)
		metric.Delta = &counter.Value
		metric.Value = nil
	default:
		err = models.ErrInvalidMetricType
	}

	if err != nil {
		s.logger.Error("failed to save metric", err)
	}

	return metric, err
}

func (s *MetricService) PutList(ctx context.Context, metrics []models.Metrics) error {
	counters := []storage.Counter{}
	gauges := make([]storage.Gauge, 0, len(metrics))
	metricNames := make([]string, 0, len(metrics))

	for _, metric := range metrics {
		switch metric.MType {
		case models.TypeGauge:
			g := storage.Gauge{Name: metric.ID, Value: *metric.Value}
			gauges = append(gauges, g)
			metricNames = append(metricNames, metric.ID)
		case models.TypeCounter:
			c := storage.Counter{Name: metric.ID, Value: *metric.Delta}
			counters = append(counters, c)
			metricNames = append(metricNames, metric.ID)
		}
	}

	err := s.storage.PutMetrics(ctx, counters, gauges)
	if err != nil {
		s.logger.Error("failed to save metrics", err)
		return err
	}

	clientIP, ok := ctx.Value(KeyClientIP).(string)
	if !ok {
		s.logger.Error("failed to get client ip", errors.New("empty ip"))
	}

	event := audit.Event{
		TS:      time.Now().Unix(),
		Metrics: metricNames,
		IP:      clientIP,
	}
	s.auditor.Rise(event)

	return nil
}
