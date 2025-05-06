package update

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EshkinKot1980/metrics/internal/server/middleware"
	"github.com/EshkinKot1980/metrics/internal/server/storage/memory"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type pathValues struct {
		mtype string
		name  string
		value string
	}

	type request struct {
		path        string
		values      pathValues
		contentType string
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
				path: "/update/counter/TestCounter/1",
				values: pathValues{
					mtype: "counter",
					name:  "TestCounter",
					value: "1",
				},
				contentType: "text/plain",
			},
			want: want{
				code: http.StatusOK,
				body: "",
			},
		},
		{
			name: "negative_counter",
			req: request{
				path: "/update/counter/TestCounter/3.14",
				values: pathValues{
					mtype: "counter",
					name:  "TestCounter",
					value: "3.14",
				},
				contentType: "text/plain",
			},
			want: want{
				code: http.StatusBadRequest,
				body: "invalid metric value, counter must be int64, given: 3.14",
			},
		},
		{
			name: "positive_gauge",
			req: request{
				path: "/update/gauge/TestGauge/3.14",
				values: pathValues{
					mtype: "gauge",
					name:  "TestGauge",
					value: "3.14",
				},
				contentType: "text/plain",
			},
			want: want{
				code: http.StatusOK,
				body: "",
			},
		},
		{
			name: "negative_gauge",
			req: request{
				path: "/update/gauge/TestGauge/wtf",
				values: pathValues{
					mtype: "gauge",
					name:  "TestGauge",
					value: "wtf",
				},
				contentType: "text/plain",
			},
			want: want{
				code: http.StatusBadRequest,
				body: "invalid metric value, gauge must be float64, given: wtf",
			},
		},
		{
			name: "negative_metric_type",
			req: request{
				path: "/update/unknown/TestUnknown/1",
				values: pathValues{
					mtype: "unknown",
					name:  "TestUnknown",
					value: "1",
				},
				contentType: "text/plain",
			},
			want: want{
				code: http.StatusBadRequest,
				body: "invalid metric type: unknown",
			},
		},
	}

	updater := memory.New()
	handler := New(updater)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, test.req.path, nil)
			req.Header.Set("content-type", test.req.contentType)
			req.SetPathValue("type", test.req.values.mtype)
			req.SetPathValue("name", test.req.values.name)
			req.SetPathValue("value", test.req.values.value)

			w := httptest.NewRecorder()
			middleware.ValidateMetric(http.HandlerFunc(handler.Update)).ServeHTTP(w, req)
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
