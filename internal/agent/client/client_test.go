package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
	"github.com/EshkinKot1980/metrics/internal/common/models"
)

func testRequest(r *http.Request) func(t *testing.T) {
	return func(t *testing.T) {
		assert.Equal(t, http.MethodPost, r.Method, "Request method")
		assert.Equal(t, ContentType, r.Header.Get("Content-Type"), "Request Content-Type header")
		assert.Equal(t, Path, r.URL.Path, "Request URL Path")

		var metric models.Metrics
		err := json.NewDecoder(r.Body).Decode(&metric)
		require.Nil(t, err, "Request Body decoding")
		err = metric.Validate()
		assert.Nil(t, err, "Metric data validation")
	}
}

func TestReport(t *testing.T) {
	server := httptest.NewServer(makeHadler(t))
	defer server.Close()

	storage := storage.New()
	initStorage(storage)
	client := New(storage, server.URL)
	client.Report()
}

func makeHadler(t *testing.T) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		name := "report_test"
		t.Run(name, testRequest(r))
	}

	return http.HandlerFunc(fn)
}

func initStorage(s *storage.MemoryStorage) {
	s.Put(
		[]agent.Counter{
			agent.Counter{Name: "TestCounter", Value: 13},
			agent.Counter{Name: "Visitors", Value: 256},
		},
		[]agent.Gauge{
			agent.Gauge{Name: "ConstE", Value: 2.71828},
			agent.Gauge{Name: "TTL", Value: 3.14e50},
		},
	)
}
