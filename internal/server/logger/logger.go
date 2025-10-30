// Модуль logger реализует логирование для сервеной части приложения.
package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger - обертка над zap.Logger.
type Logger struct {
	logger *zap.Logger
}

// Данные HTTP запроса.
type RequestLogData struct {
	URI      string
	Method   string
	Duration time.Duration
}

// Данные HTTP ответа.
type ResponseLogData struct {
	Status int
	Size   int
}

func New() (*Logger, error) {
	l, err := zap.NewDevelopment(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{logger: l}, nil
}

// Синхронизация логера при завершении приложения.
func (l *Logger) Sync() {
	l.logger.Sync()
}

// Логирование ошибки.
func (l *Logger) Error(message string, err error) {
	l.logger.Error(message, zap.Error(err))
}

// Логирование HTTP запроса.
func (l *Logger) RequestInfo(message string, req *RequestLogData, resp *ResponseLogData) {
	l.logger.Info(message, zap.Object("request", req), zap.Object("response", resp))
}

// Преобразует данные запроса в zap.Field.
func (o *RequestLogData) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("uri", o.URI)
	enc.AddString("method", o.Method)
	enc.AddDuration("duration", o.Duration)
	return nil
}

// Преобразует данные ответа в zap.Field.
func (o *ResponseLogData) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("status", o.Status)
	enc.AddInt("size", o.Size)
	return nil
}
