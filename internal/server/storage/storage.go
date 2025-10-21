// Модуль storage реализует хранение данных на сервере.
package storage

import (
	"context"
	"errors"

	"github.com/EshkinKot1980/metrics/internal/server/config"
)

// Ошибки возвращемые, если метрика не найдена.
var (
	ErrCounterNotFound = errors.New("counter metric not found")
	ErrGaugeNotFound   = errors.New("gauge metric not found")
)

// Модель метрики - тип датчик.
type Gauge struct {
	Name  string
	Value float64
}

// Модель метрики - тип счетчик.
type Counter struct {
	Name  string
	Value int64
}

// Storage определяет интерфейс для хранилища метрик и операций с ними.
type Storage interface {
	// Получение счетчика, в качестве ошики может вернуть ErrCounterNotFound.
	GetCounter(c Counter) (Counter, error)
	// Получение датчика, в качестве ошики может вернуть ErrGaugeNotFound.
	GetGauge(g Gauge) (Gauge, error)
	// Сохранение счетчика, возвращает его обновленне значение.
	PutCounter(c Counter) (Counter, error)
	// Сохранение датчика.
	PutGauge(g Gauge) error
	// Множественное сохранение метрик.
	PutMetrics(ctx context.Context, counters []Counter, gauges []Gauge) error
	// Выключение хранилища.
	Halt()
	// Проверка связи с базой данных.
	Ping() bool
}

// Loogger, логирует ошибки.
type Logger interface {
	Error(message string, err error)
}

// Фабричный метод для создания хранилища.
func New(cfg *config.Config, l Logger) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		return NewDBStorage(cfg.DatabaseDSN)
	}

	return NewFileStorage(cfg.FileCfg, l)
}
