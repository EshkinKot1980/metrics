// Модуль service реализует сервисный слой для сервера сбора метрик.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/EshkinKot1980/metrics/internal/common/models"
	"github.com/EshkinKot1980/metrics/internal/server/audit"
	"github.com/EshkinKot1980/metrics/internal/server/storage"
)

// Ключ для хранения данных в контексте
type ContextKey string

// IPv4 адрес клиента.
const KeyClientIP ContextKey = "clientIP"

// Метрика отсутсвует в хранилище.
var ErrMetricNotFound = errors.New("metric not found")

// Хранилище метрик.
type Storage interface {
	GetCounter(c storage.Counter) (storage.Counter, error)
	GetGauge(g storage.Gauge) (storage.Gauge, error)
	PutCounter(c storage.Counter) (storage.Counter, error)
	PutGauge(g storage.Gauge) error
	PutMetrics(ctx context.Context, counters []storage.Counter, gauges []storage.Gauge) error
}

// Loogger, логирует ошибки.
type Logger interface {
	Error(message string, err error)
}

// Auditor осуществляет аудит сохраненных метрик.
type Auditor interface {
	// Создает событие аудита.
	Rise(e audit.Event)
}

// Сервис для работы с метриками.
type MetricService struct {
	storage Storage
	logger  Logger
	auditor Auditor
}

func NewMetricService(s Storage, l Logger, a Auditor) *MetricService {
	return &MetricService{storage: s, logger: l, auditor: a}
}

// Заполняет метрику из хранилища. Получает метрику заполнеными полями ID и MType,
// возвращает её с заполнеными полями Value или Delta в зависимости от типа.
// В качестве ошибки может вернуть ErrMetricNotFound.
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

// Сохраняет метрику в хранилище, возвращает её обновленное состояние.
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

// Сохраняет множество метрик в хранилище.
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
