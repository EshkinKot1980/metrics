package http

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rsa"
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
	"github.com/EshkinKot1980/metrics/internal/common/utils"
)

func testBatchRequest(r *http.Request, priv *rsa.PrivateKey) func(t *testing.T) {
	return func(t *testing.T) {
		assert.Equal(t, http.MethodPost, r.Method, "Request method")
		assert.Equal(t, BatchPath, r.URL.Path, "Request URL Path")
		assert.Equal(t, ContentType, r.Header.Get("Content-Type"), "Request Content-Type header")
		assert.Contains(t, r.Header.Get("Accept-Encoding"), "gzip", "Request Accept-Encoding header")
		assert.Contains(t, r.Header.Get("Content-Encoding"), "gzip", "Request Content-Encoding header")
		if priv != nil {
			assert.Contains(t, r.Header.Get("X-Encrypted"), "true", "Request Content-Encoding header")
		}

		ip, err := defineIP()
		require.Nil(t, err, "Define IP address")
		assert.Equal(t, ip, r.Header.Get("X-Real-IP"), "Request X-Real-IP header")

		gz, err := gzip.NewReader(r.Body)
		require.Nil(t, err, "Request Body decompressing: creating reader)")
		defer gz.Close()

		body, err := io.ReadAll(gz)
		require.Nil(t, err, "Request Body decompressing: reading body")

		if priv != nil {
			body, err = utils.DecryptWithPrivateKey(body, priv)
			require.Nil(t, err, "Request Body decrypting")
		}

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

func TestBathcClient_Report(t *testing.T) {
	priv, pub, err := utils.GenerateKeyPair()
	require.Nil(t, err, "Generate rsa key pair")

	tests := []struct {
		name string
		pub  *rsa.PublicKey
		priv *rsa.PrivateKey
	}{
		{
			name: "not_encrypted",
		},
		{
			name: "encrypted",
			pub:  pub,
			priv: priv,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				testBatchRequest(r, test.priv)
			})

			server := httptest.NewServer(handler)
			defer server.Close()

			storage := storage.New()
			testInitStorage(storage)
			client, err := NewBatchClient(storage, server.URL, "secret", test.pub)
			require.Nil(t, err, "Init client")
			client.Report()
		})
	}

}

func testInitStorage(s *storage.MemoryStorage) {
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
