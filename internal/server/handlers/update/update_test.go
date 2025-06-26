package update

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/EshkinKot1980/metrics/internal/server/storage/memory"
)

func TestUpdateHandler(t *testing.T) {
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
				body: "invalid metric value, counter must be int64, given: 3.14",
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
				body: "invalid metric value, gauge must be float64, given: wtf",
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

	updater := memory.New()
	logger := LoggerStub{}
	handler := New(updater, logger)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, test.path, nil)
			r.Header.Set("content-type", "text/plain")
			r.SetPathValue("type", test.values.mtype)
			r.SetPathValue("name", test.values.name)
			r.SetPathValue("value", test.values.value)

			w := httptest.NewRecorder()
			handler.Update(w, r)
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
