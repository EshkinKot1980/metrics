package storage

import "errors"

var (
	ErrCounterNotFound  = errors.New("counter metric not found")
	ErrGaugeNotFound    = errors.New("cauge metric not found")
)
