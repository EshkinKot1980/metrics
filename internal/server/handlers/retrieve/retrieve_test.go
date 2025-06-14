package retrieve

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/EshkinKot1980/metrics/internal/server/storage/memory"
)

func TestNew(t *testing.T) {
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
				body: "counter metric not found",
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
				body: "gauge metric not found",
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
				code: http.StatusNotFound,
				body: "invalid metric type: unknown",
			},
		},
	}

	retriever := memory.New()
	retriever.PutCounter("TestCounter", 13)
	retriever.PutGauge("TestGauge", 3.14)
	logger := LoggerStub{}
	handler := New(retriever, logger)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, test.req.path, nil)
			req.SetPathValue("type", test.req.values.mtype)
			req.SetPathValue("name", test.req.values.name)

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

type LoggerStub struct{}

func (l LoggerStub) Error(message string, err error) {}
