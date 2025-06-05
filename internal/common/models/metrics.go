package models

import "errors"

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

var (
	ErrInvalidGauge      = errors.New("gauge metric must contain float64 \"value\" field")
	ErrIvalidCounter     = errors.New("counter metric must contain int64 \"delta\" field")
	ErrInvalidMetricType = errors.New("invalid metric type")
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// Проверяет входящие на сервер данные
func (m Metrics) Validate() error {
	switch m.MType {
	case TypeGauge:
		if m.Value == nil {
			return ErrInvalidGauge
		}
	case TypeCounter:
		if m.Delta == nil {
			return ErrIvalidCounter
		}
	default:
		return ErrInvalidMetricType
	}

	return nil
}
