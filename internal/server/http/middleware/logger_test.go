package middleware

import (
	// "bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPLogger(t *testing.T) {
	lm := makeLoggerMock(t)
	mw := NewHTTPLogger(lm)
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "four")
	})

	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()
	handler := mw.Log(next)
	handler.ServeHTTP(w, r)
}

type loggerMock struct {
	t *testing.T
}

func makeLoggerMock(t *testing.T) loggerMock {
	return loggerMock{t: t}
}

func (l loggerMock) RequestInfo(message string, req *requestData, resp *responseData) {
	assert.Equal(l.t, "server api", message, "Log message")
	assert.Equal(l.t, "/path", req.URI, "Log request URI")
	assert.Equal(l.t, http.MethodGet, req.Method, "Log request method")
	assert.NotEmpty(l.t, req.Duration, "Log request duration")
	assert.Equal(l.t, http.StatusNotFound, resp.Status, "Log response status code")
	assert.Equal(l.t, 4, resp.Size, "Log response content length")
}
