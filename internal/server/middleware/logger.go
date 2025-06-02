package middleware

import (
	"net/http"
	"time"

	"github.com/EshkinKot1980/metrics/internal/server"
)

type HTTPLogger struct {
	logger HTTPLogWriter
}

type requestData = server.RequestLogData
type responseData = server.ResponseLogData

type HTTPLogWriter interface {
	RequestInfo(message string, req *requestData, resp *responseData)
}

func NewHTTPLogger(l HTTPLogWriter) *HTTPLogger {
	return &HTTPLogger{logger: l}
}

func (h *HTTPLogger) Log(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			Status: 0,
			Size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r)

		requestData := &requestData{
			URI:      r.RequestURI,
			Method:   r.Method,
			Duration: time.Since(start),
		}

		h.logger.RequestInfo("server api", requestData, responseData)
	}

	return http.HandlerFunc(fn)
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.Size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.Status = statusCode
}
