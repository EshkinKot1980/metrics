package storage

import "errors"

var (
	ErrCounterNotFound = errors.New("counter metric not found")
	ErrGaugeNotFound   = errors.New("gauge metric not found")
)

type Gauge struct {
	ID    string
	Value float64
}

type Counter struct {
	ID    string
	Delta int64
}
