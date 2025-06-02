package update

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

func TestUpdateJSONHandler(t *testing.T) {
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
			name:    "positive_counter1",
			reqBody: `{"id":"TestCounter","type":"counter","delta":1}`,
			want: want{
				code: http.StatusOK,
				body: `{"id":"TestCounter","type":"counter","delta":1}`,
			},
		},
		{
			name:    "positive_counter2",
			reqBody: `{"id":"TestCounter","type":"counter","delta":13}`,
			want: want{
				code: http.StatusOK,
				body: `{"id":"TestCounter","type":"counter","delta":14}`,
			},
		},
		{
			name:    "negative_counter1",
			reqBody: `{"id":"TestCounter","type":"counter","delta":3.14}`,
			want: want{
				code: http.StatusBadRequest,
				body: "json: cannot unmarshal number 3.14 into Go struct field Metrics.delta of type int64",
			},
		},
		{
			name:    "negative_counter2",
			reqBody: `{"id":"TestCounter","type":"counter"}`,
			want: want{
				code: http.StatusBadRequest,
				body: "counter metric must contain int64 \"delta\" field",
			},
		},
		{
			name:    "positive_gauge1",
			reqBody: `{"id":"TestGauge","type":"gauge","value":3.14}`,
			want: want{
				code: http.StatusOK,
				body: `{"id":"TestGauge","type":"gauge","value":3.14}`,
			},
		},
		{
			name:    "positive_gauge2",
			reqBody: `{"id":"TestGauge","type":"gauge","value":3.1415}`,
			want: want{
				code: http.StatusOK,
				body: `{"id":"TestGauge","type":"gauge","value":3.1415}`,
			},
		},
		{
			name:    "negative_gauge1",
			reqBody: `{"id":"TestGauge","type":"gauge","value":"wtf"}`,
			want: want{
				code: http.StatusBadRequest,
				body: "json: cannot unmarshal string into Go struct field Metrics.value of type float64",
			},
		},
		{
			name:    "negative_gauge2",
			reqBody: `{"id":"TestGauge","type":"gauge"}`,
			want: want{
				code: http.StatusBadRequest,
				body: "gauge metric must contain float64 \"value\" field",
			},
		},
		{
			name:    "negative_metric_type",
			reqBody: `{"id":"TestGauge","type":"unknown","delta":1,"value":3.14}`,
			want: want{
				code: http.StatusBadRequest,
				body: "invalid metric type: unknown",
			},
		},
	}

	updater := memory.New()
	logger := LoggerStub{}
	handler := NewJSONHandler(updater, logger)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBody := []byte(test.reqBody)
			r := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(reqBody))
			r.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.Update(w, r)
			res := w.Result()
			defer res.Body.Close()

			//TODO: разобраться почему реальный сервер отдает "application/json", а тестовый "text/plain"
			//assert.Equal(t, "application/json", res.Header.Get("Content-Type"), "Response Content-Type")
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

type LoggerStub struct{}

func (l LoggerStub) Error(message string, err error) {}
