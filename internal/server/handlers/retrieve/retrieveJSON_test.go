package retrieve

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/EshkinKot1980/metrics/internal/server/storage/memory"
)

func TestValueJSONHandler(t *testing.T) {
	type want struct {
		code int
		body string
	}

	tests := []struct {
		name    string
		reqBody string
		want    want
	}{
		{
			name:    "positive_counter",
			reqBody: `{"id":"TestCounter","type":"counter"}`,
			want: want{
				code: http.StatusOK,
				body: `{"id":"TestCounter","type":"counter","delta":13}`,
			},
		},
		{
			name:    "negative_counter",
			reqBody: `{"id":"Unknown","type":"counter"}`,
			want: want{
				code: http.StatusNotFound,
				body: "counter metric not found",
			},
		},
		{
			name:    "positive_gauge",
			reqBody: `{"id":"TestGauge","type":"gauge"}`,
			want: want{
				code: http.StatusOK,
				body: `{"id":"TestGauge","type":"gauge","value":3.14}`,
			},
		},
		{
			name:    "negative_gauge",
			reqBody: `{"id":"Unknown","type":"gauge"}`,
			want: want{
				code: http.StatusNotFound,
				body: "gauge metric not found",
			},
		},
		{
			name:    "negative_metric_type",
			reqBody: `{"id":"TestUnknown","type":"unknown"}`,
			want: want{
				code: http.StatusNotFound,
				body: "invalid metric type: unknown",
			},
		},
	}

	retriever := memory.New()
	retriever.PutCounter("TestCounter", 13)
	retriever.PutGauge("TestGauge", 3.14)
	logger := LoggerStub{}
	handler := NewJSONHandler(retriever, logger)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBody := []byte(test.reqBody)
			req := httptest.NewRequest(http.MethodGet, "/value", bytes.NewBuffer(reqBody))

			w := httptest.NewRecorder()
			handler.Retrieve(w, req)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode, "Response status code")
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			body := strings.TrimSuffix(string(resBody), "\n")
			assert.Equal(t, test.want.body, body, "Response body")
		})
	}
}
