package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EshkinKot1980/metrics/internal/server/service"
	"github.com/EshkinKot1980/metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHandler_UpdateFromPath(t *testing.T) {
	type pathValues struct {
		mtype string
		name  string
		value string
	}

	type want struct {
		code int
		body string
	}

	tests := []struct {
		name   string
		path   string
		values pathValues
		want   want
	}{
		{
			name: "positive_counter",
			path: "/update/counter/TestCounter/1",
			values: pathValues{
				mtype: "counter",
				name:  "TestCounter",
				value: "1",
			},
			want: want{
				code: http.StatusOK,
				body: "",
			},
		},
		{
			name: "negative_counter",
			path: "/update/counter/TestCounter/3.14",
			values: pathValues{
				mtype: "counter",
				name:  "TestCounter",
				value: "3.14",
			},
			want: want{
				code: http.StatusBadRequest,
				body: "invalid counter, value must be int64, given: 3.14",
			},
		},
		{
			name: "positive_gauge",
			path: "/update/gauge/TestGauge/3.14",
			values: pathValues{
				mtype: "gauge",
				name:  "TestGauge",
				value: "3.14",
			},
			want: want{
				code: http.StatusOK,
				body: "",
			},
		},
		{
			name: "negative_gauge",
			path: "/update/gauge/TestGauge/wtf",
			values: pathValues{
				mtype: "gauge",
				name:  "TestGauge",
				value: "wtf",
			},
			want: want{
				code: http.StatusBadRequest,
				body: "invalid gauge, value must be float64, given: wtf",
			},
		},
		{
			name: "negative_metric_type",
			path: "/update/unknown/TestUnknown/1",
			values: pathValues{
				mtype: "unknown",
				name:  "TestUnknown",
				value: "1",
			},
			want: want{
				code: http.StatusBadRequest,
				body: "invalid metric type: unknown",
			},
		},
	}

	s := storage.NewMemoryStorage()
	logger := LoggerStub{}
	srv := service.NewMetricService(s, logger)
	handler := NewUpdateHandler(srv, logger)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, test.path, nil)
			r.Header.Set("content-type", "text/plain")
			r.SetPathValue("type", test.values.mtype)
			r.SetPathValue("name", test.values.name)
			r.SetPathValue("value", test.values.value)

			w := httptest.NewRecorder()
			handler.UpdateFromPath(w, r)
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

func TestUpdateHandler_Update(t *testing.T) {
	type want struct {
		code        int
		contentType string
		body        string
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
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id":"TestCounter","type":"counter","delta":1}`,
			},
		},
		{
			name:    "positive_counter2",
			reqBody: `{"id":"TestCounter","type":"counter","delta":13}`,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id":"TestCounter","type":"counter","delta":14}`,
			},
		},
		{
			name:    "negative_counter1",
			reqBody: `{"id":"TestCounter","type":"counter","delta":3.14}`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "json: cannot unmarshal number 3.14 into Go struct field Metrics.delta of type int64",
			},
		},
		{
			name:    "negative_counter2",
			reqBody: `{"id":"TestCounter","type":"counter"}`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "invalid counter, metric must contain int64 \"delta\" field",
			},
		},
		{
			name:    "positive_gauge1",
			reqBody: `{"id":"TestGauge","type":"gauge","value":3.14}`,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id":"TestGauge","type":"gauge","value":3.14}`,
			},
		},
		{
			name:    "positive_gauge2",
			reqBody: `{"id":"TestGauge","type":"gauge","value":3.1415}`,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id":"TestGauge","type":"gauge","value":3.1415}`,
			},
		},
		{
			name:    "negative_gauge1",
			reqBody: `{"id":"TestGauge","type":"gauge","value":"wtf"}`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "json: cannot unmarshal string into Go struct field Metrics.value of type float64",
			},
		},
		{
			name:    "negative_gauge2",
			reqBody: `{"id":"TestGauge","type":"gauge"}`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "invalid gauge, metric must contain float64 \"value\" field",
			},
		},
		{
			name:    "negative_metric_type",
			reqBody: `{"id":"TestGauge","type":"unknown","delta":1,"value":3.14}`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "invalid metric type",
			},
		},
	}

	s := storage.NewMemoryStorage()
	logger := LoggerStub{}
	srv := service.NewMetricService(s, logger)
	handler := NewUpdateHandler(srv, logger)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBody := []byte(test.reqBody)
			r := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(reqBody))
			r.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.Update(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"), "Response Content-Type")
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

func TestUpdateHandler_UpdateList(t *testing.T) {
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
			name: "positive_data",
			reqBody: `[
						{"id":"TestCounter","type":"counter","delta":1},
						{"id":"TestGauge","type":"gauge","value":3.14}
					  ]`,
			want: want{
				code: http.StatusOK,
				body: "",
			},
		},
		{
			name: "negative_counter1",
			reqBody: `[
						{"id":"ValidCounter","type":"counter","delta":1},
						{"id":"TestCounter","type":"counter","delta":3.14}
					  ]`,
			want: want{
				code: http.StatusBadRequest,
				body: "json: cannot unmarshal number 3.14 into Go struct field Metrics.delta of type int64",
			},
		},
		{
			name: "negative_counter2",
			reqBody: `[
						{"id":"ValidCounter","type":"counter","delta":1},
						{"id":"TestCounter","type":"counter"}
					  ]`,
			want: want{
				code: http.StatusBadRequest,
				body: "invalid counter, metric must contain int64 \"delta\" field {id: TestCounter, type: counter}",
			},
		},
		{
			name: "negative_gauge1",
			reqBody: `[
						{"id":"ValidCounter","type":"counter","delta":1},
						{"id":"TestGauge","type":"gauge","value":"wtf"}
					  ]`,
			want: want{
				code: http.StatusBadRequest,
				body: "json: cannot unmarshal string into Go struct field Metrics.value of type float64",
			},
		},
		{
			name: "negative_gauge2",
			reqBody: `[
						{"id":"ValidCounter","type":"counter","delta":1},
						{"id":"TestGauge","type":"gauge"}
					  ]`,
			want: want{
				code: http.StatusBadRequest,
				body: "invalid gauge, metric must contain float64 \"value\" field {id: TestGauge, type: gauge}",
			},
		},
		{
			name: "negative_metric_type",
			reqBody: `[
						{"id":"ValidCounter","type":"counter","delta":1},
						{"id":"TestGauge","type":"unknown","delta":1,"value":3.14}
					  ]`,
			want: want{
				code: http.StatusBadRequest,
				body: "invalid metric type {id: TestGauge, type: unknown}",
			},
		},
	}

	s := storage.NewMemoryStorage()
	logger := LoggerStub{}
	srv := service.NewMetricService(s, logger)
	handler := NewUpdateHandler(srv, logger)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBody := []byte(test.reqBody)
			r := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(reqBody))
			r.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.UpdateList(w, r)
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
