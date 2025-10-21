package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/EshkinKot1980/metrics/internal/server/audit"
	"github.com/EshkinKot1980/metrics/internal/server/service"
	"github.com/EshkinKot1980/metrics/internal/server/storage"
)

func TestRetrieveHandler_GetByPath(t *testing.T) {
	type pathValues struct {
		mtype string
		name  string
	}

	type request struct {
		path   string
		values pathValues
	}

	type want struct {
		code int
		body string
	}

	tests := []struct {
		name string
		req  request
		want want
	}{
		{
			name: "positive_counter",
			req: request{
				path: "/value/counter/TestCounter",
				values: pathValues{
					mtype: "counter",
					name:  "TestCounter",
				},
			},
			want: want{
				code: http.StatusOK,
				body: "13",
			},
		},
		{
			name: "negative_counter",
			req: request{
				path: "/value/counter/Unknown",
				values: pathValues{
					mtype: "counter",
					name:  "Unknown",
				},
			},
			want: want{
				code: http.StatusNotFound,
				body: "metric not found",
			},
		},
		{
			name: "positive_gauge",
			req: request{
				path: "/value/gauge/TestGauge",
				values: pathValues{
					mtype: "gauge",
					name:  "TestGauge",
				},
			},
			want: want{
				code: http.StatusOK,
				body: "3.14",
			},
		},
		{
			name: "negative_gauge",
			req: request{
				path: "/value/gauge/Unknown",
				values: pathValues{
					mtype: "gauge",
					name:  "Unknown",
				},
			},
			want: want{
				code: http.StatusNotFound,
				body: "metric not found",
			},
		},
		{
			name: "negative_metric_type",
			req: request{
				path: "/value/unknown/TestUnknown",
				values: pathValues{
					mtype: "unknown",
					name:  "TestUnknown",
				},
			},
			want: want{
				code: http.StatusBadRequest,
				body: "invalid metric type",
			},
		},
	}

	s := storage.NewMemoryStorage()
	s.PutCounter(storage.Counter{Name: "TestCounter", Value: 13})
	s.PutGauge(storage.Gauge{Name: "TestGauge", Value: 3.14})
	logger := LoggerStub{}
	auditor := AuditorStub{}
	srv := service.NewMetricService(s, logger, auditor)
	handler := NewRetrieveHandler(srv, logger)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, test.req.path, nil)
			req.SetPathValue("type", test.req.values.mtype)
			req.SetPathValue("name", test.req.values.name)

			w := httptest.NewRecorder()
			handler.GetByPath(w, req)
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

func TestRetrieveHandler_GetJSON(t *testing.T) {
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
				body: "metric not found",
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
				body: "metric not found",
			},
		},
		{
			name:    "negative_metric_type",
			reqBody: `{"id":"TestUnknown","type":"unknown"}`,
			want: want{
				code: http.StatusBadRequest,
				body: "invalid metric type",
			},
		},
	}

	s := storage.NewMemoryStorage()
	s.PutCounter(storage.Counter{Name: "TestCounter", Value: 13})
	s.PutGauge(storage.Gauge{Name: "TestGauge", Value: 3.14})
	logger := LoggerStub{}
	auditor := AuditorStub{}
	srv := service.NewMetricService(s, logger, auditor)
	handler := NewRetrieveHandler(srv, logger)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBody := []byte(test.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(reqBody))

			w := httptest.NewRecorder()
			handler.GetJSON(w, req)
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

type LoggerStub struct{}

func (l LoggerStub) Error(message string, err error) {}

type AuditorStub struct{}

func (a AuditorStub) Rise(e audit.Event) {}
