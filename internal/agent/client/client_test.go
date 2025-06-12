package client

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
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
		assert.Equal(t, Path, r.URL.Path, "Request URL Path")
		assert.Equal(t, ContentType, r.Header.Get("Content-Type"), "Request Content-Type header")
		assert.Contains(t, r.Header.Get("Accept-Encoding"), "gzip", "Request Accept-Encoding header")
		assert.Contains(t, r.Header.Get("Content-Encoding"), "gzip", "Request Content-Encoding header")

		gz, err := gzip.NewReader(r.Body)
		require.Nil(t, err, "Request Body decompressing: creating reader)")
		defer gz.Close()

		body, err := io.ReadAll(gz)
		require.Nil(t, err, "Request Body decompressing: reading body")

		var metrics []models.Metrics
		bodyReader := bytes.NewReader(body)
		err = json.NewDecoder(bodyReader).Decode(&metrics)
		require.Nil(t, err, "Request Body decoding")

		for _, m := range metrics {
			assert.Nil(t, m.Validate(), "Metric "+m.ID+" data validation")
		}

		requestHash := r.Header.Get("HashSHA256")
		assert.NotEmpty(t, requestHash, "Request HashSHA256 header: notempty")

		h := hmac.New(sha256.New, []byte("secret"))
		h.Write(body)
		wantedHash := hex.EncodeToString(h.Sum(nil))
		assert.Equal(t, wantedHash, requestHash, "Request HashSHA256 header: check")
	}
}

func TestReport(t *testing.T) {
	server := httptest.NewServer(makeHadler(t))
	defer server.Close()

	storage := storage.New()
	initStorage(storage)
	client := New(storage, server.URL, "secret")
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
