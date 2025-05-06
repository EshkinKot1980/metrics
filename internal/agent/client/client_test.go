package client

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EshkinKot1980/metrics/internal/agent/model"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
)

func testRequest(req *http.Request) func(t *testing.T) {
	return func(t *testing.T) {
		assert.Equal(t, http.MethodPost, req.Method, "Request method")
		assert.Equal(t, ContentType, req.Header.Get("Content-Type"), "Request Content-Type header")

		pathParts := strings.Split(req.URL.Path, "/")
		require.Equal(t, 5, len(pathParts), "Split path count")
		assert.Equal(t, PathPrefix, pathParts[1], "Path prefix")
		require.Contains(t, []string{TypeCounter, TypeGauge}, pathParts[2], "Metric type")

		switch pathParts[2] {
		case TypeGauge:
			_, err := strconv.ParseFloat(pathParts[4], 64)
			assert.Nil(t, err, "Check gauge value")
		case TypeCounter:
			_, err := strconv.ParseInt(pathParts[4], 10, 64)
			assert.Nil(t, err, "Check counter value")
		}
	}
}

func TestReport(t *testing.T) {
	server := httptest.NewServer(makeHadler(t))
	defer server.Close()

	st := storage.New()
	initStorage(st)
	c := New(st, server.URL)
	c.Report()
}

func makeHadler(t *testing.T) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		name := "report_path:" + req.URL.Path
		t.Run(name, testRequest(req))
	}

	return http.HandlerFunc(fn)
}

func initStorage(s *storage.MemoryStorage) {
	s.Put(
		[]model.Counter{
			model.Counter{Name: "TestCounter", Value: 13},
			model.Counter{Name: "Visitors", Value: 256},
		},
		[]model.Gauge{
			model.Gauge{Name: "ConstE", Value: 2.71828},
			model.Gauge{Name: "TTL", Value: 3.14e50},
		},
	)
}
