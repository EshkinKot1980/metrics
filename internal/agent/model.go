// Модуль agent реализует клиентскую часть приложения сбора метрик.
package agent

// Модель датчика.
type Gauge struct {
	Name  string
	Value float64
}

// Модель счетчика.
type Counter struct {
	Name  string
	Value int64
}
