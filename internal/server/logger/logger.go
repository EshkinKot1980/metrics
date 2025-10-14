package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	logger *zap.Logger
}

type RequestLogData struct {
	URI      string
	Method   string
	Duration time.Duration
}

type ResponseLogData struct {
	Status int
	Size   int
}

// В дальнейшем здесь будет конфигурация зависящая от окружения (environment)
func New() (*Logger, error) {
	l, err := zap.NewDevelopment(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{logger: l}, nil
}

func (l *Logger) Sync() {
	l.logger.Sync()
}

func (l *Logger) Error(message string, err error) {
	l.logger.Error(message, zap.Error(err))
}

func (l *Logger) RequestInfo(message string, req *RequestLogData, resp *ResponseLogData) {
	l.logger.Info(message, zap.Object("request", req), zap.Object("response", resp))
}

func (o *RequestLogData) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("uri", o.URI)
	enc.AddString("method", o.Method)
	enc.AddDuration("duration", o.Duration)
	return nil
}

func (o *ResponseLogData) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("status", o.Status)
	enc.AddInt("size", o.Size)
	return nil
}
