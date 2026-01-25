package http

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

	"github.com/EshkinKot1980/metrics/internal/agent/storage"
	"github.com/EshkinKot1980/metrics/internal/common/models"
)

var testCounterChan = make(chan int)

func testSingleRequest(r *http.Request) func(t *testing.T) {
	return func(t *testing.T) {
		assert.Equal(t, http.MethodPost, r.Method, "Request method")
		assert.Equal(t, ContentType, r.Header.Get("Content-Type"), "Request Content-Type header")
		assert.Equal(t, SinglePath, r.URL.Path, "Request URL Path")

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

	t.Run("query_count", func(t *testing.T) {
		assert.Equal(t, 4, count, "Requests count")
	})
}

func TestMultithreadedClient_Report(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Run("request", testSingleRequest(r))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	storage := storage.New()
	testInitStorage(storage)
	client := NewMultithreadedClient(storage, server.URL, 2)
	client.Report()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second/10)
	defer cancel()

	testQueryCount(ctx, t)
}
