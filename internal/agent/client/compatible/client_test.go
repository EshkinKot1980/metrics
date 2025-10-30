package compatible

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
	"github.com/EshkinKot1980/metrics/internal/common/models"
)

var testCounterChan = make(chan int)

func testRequest(r *http.Request) func(t *testing.T) {
	return func(t *testing.T) {
		assert.Equal(t, http.MethodPost, r.Method, "Request method")
		assert.Equal(t, ContentType, r.Header.Get("Content-Type"), "Request Content-Type header")
		assert.Equal(t, Path, r.URL.Path, "Request URL Path")

		gz, err := gzip.NewReader(r.Body)
		require.Nil(t, err, "Request Body decompressing: creating reader)")
		defer gz.Close()

		body, err := io.ReadAll(gz)
		require.Nil(t, err, "Request Body decompressing: reading body")

		var metric models.Metrics
		bodyReader := bytes.NewReader(body)
		err = json.NewDecoder(bodyReader).Decode(&metric)
		require.Nil(t, err, "Request Body decoding")

		err = metric.Validate()
		assert.Nil(t, err, "Metric data validation")

		testCounterChan <- 1
	}
}

func testQueryCount(ctx context.Context, t *testing.T) {
	count := 0
done:
	for {
		select {
		case <-ctx.Done():
			break done
		case <-testCounterChan:
			count++
		}

		if count >= 4 {
			break done
		}
	}

	t.Run("query_count_test", func(t *testing.T) {
		assert.Equal(t, 4, count, "Requests count")
	})
}

func TestReport(t *testing.T) {
	server := httptest.NewServer(makeHadler(t))
	defer server.Close()

	storage := storage.New()
	initStorage(storage)
	client := New(storage, server.URL, 2)
	client.Report()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second/10)
	defer cancel()

	testQueryCount(ctx, t)
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
			{Name: "TestCounter", Value: 13},
			{Name: "Visitors", Value: 256},
		},
		[]agent.Gauge{
			{Name: "ConstE", Value: 2.71828},
			{Name: "TTL", Value: 3.14e50},
		},
	)
}
