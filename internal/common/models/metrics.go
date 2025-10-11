package models

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
	IDmaxLen    = 32
)

var (
	ErrInvalidGauge      = errors.New("invalid gauge")
	ErrIvalidCounter     = errors.New("invalid counter")
	ErrInvalidMetricType = errors.New("invalid metric type")
	ErrIDisTooLong       = errors.New("id is too long, maximum 32 characters")
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func MakeMetrics(id, mType, value string) (Metrics, error) {
	metric := Metrics{
		ID:    id,
		MType: mType,
	}

	if len(metric.ID) > IDmaxLen {
		return metric, ErrIDisTooLong
	}

	switch metric.MType {
	case TypeGauge:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return metric, fmt.Errorf("%w, value must be float64, given: %s", ErrInvalidGauge, value)
		}
		metric.Value = &v
	case TypeCounter:
		d, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return metric, fmt.Errorf("%w, value must be int64, given: %s", ErrIvalidCounter, value)
		}
		metric.Delta = &d
	default:
		return metric, fmt.Errorf("%w: %s", ErrInvalidMetricType, metric.MType)
	}

	return metric, nil
}

// Проверяет входящие на сервер данные
func (m Metrics) Validate() error {
	switch m.MType {
	case TypeGauge:
		if m.Value == nil {
			return fmt.Errorf("%w, metric must contain float64 \"value\" field", ErrInvalidGauge)
		}
	case TypeCounter:
		if m.Delta == nil {
			return fmt.Errorf("%w, metric must contain int64 \"delta\" field", ErrIvalidCounter)
		}
	default:
		return ErrInvalidMetricType
	}

	if len(m.ID) > IDmaxLen {
		return ErrIDisTooLong
	}
	return nil
}

func (m Metrics) ValidateType() error {
	switch m.MType {
	case TypeGauge, TypeCounter:
		return nil
	default:
		return ErrInvalidMetricType
	}
}
