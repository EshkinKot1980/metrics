package storage

import "errors"

var (
	ErrCounterNotFound = errors.New("counter metric not found")
	ErrGaugeNotFound   = errors.New("gauge metric not found")
)
