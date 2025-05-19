package models

import "fmt"

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
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
			return fmt.Errorf("gauge metric must contain float64 \"value\" field")
		}
	case TypeCounter:
		if m.Delta == nil {
			return fmt.Errorf("counter metric must contain int64 \"delta\" field")
		}
	default:
		return fmt.Errorf("invalid metric type: %s", m.MType)
	}

	return nil
}
