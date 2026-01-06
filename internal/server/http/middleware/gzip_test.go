package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipWrapperDecompress(t *testing.T) {
	var compressedBody bytes.Buffer
	rawBody := []byte(`{"data": "some data"}`)

	g := gzip.NewWriter(&compressedBody)
	_, err := g.Write(rawBody)
	require.Nil(t, err, "Creating gzip writer")
	require.Nil(t, g.Close(), "Close gzip writer")

	type want struct {
		code int
		body string
	}

	tests := []struct {
		name       string
		sendHeader bool
		reqBody    io.Reader
		want       want
	}{
		{
			name:       "positive_compressed",
			sendHeader: true,
			reqBody:    &compressedBody,
			want: want{
				code: http.StatusOK,
				body: string(rawBody),
			},
		},
		{
			name:       "positive_not_compressed",
			sendHeader: false,
			reqBody:    bytes.NewBuffer(rawBody),
			want: want{
				code: http.StatusOK,
				body: string(rawBody),
			},
		},
		{
			name:       "negative_bad_body",
			sendHeader: true,
			reqBody:    bytes.NewBuffer(rawBody),
			want: want{
				code: http.StatusBadRequest,
				body: "failed to decompress request body",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", test.reqBody)
			if test.sendHeader {
				r.Header.Set("Content-Encoding", "gzip")
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				body, err := io.ReadAll(req.Body)
				require.Nil(t, err, "Request body read")
				assert.Equal(t, test.want.body, string(body), "Handler request body")
			})

			w := httptest.NewRecorder()
			handler := GzipWrapper(next)
			handler.ServeHTTP(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode, "Response status code")
		})
	}
}

func testBodyCompress(t *testing.T, body string) string {
	var compressed bytes.Buffer
	t.Log("Compress body: ", body)

	g := gzip.NewWriter(&compressed)
	_, err := g.Write([]byte(body))
	require.Nil(t, err, "Creating gzip writer")
	require.Nil(t, g.Close(), "Close gzip writer")

	res, err := io.ReadAll(&compressed)
	require.Nil(t, err, "Read compressed body")
	return string(res)
}

func TestGzipWrapperCompress(t *testing.T) {
	bodyJSON := `{"data": "some data"}`
	bodyHTML := "<html><body>Test HTML body</body></html>"
	bodyJSONgz := testBodyCompress(t, bodyJSON)
	bodyHTMLgz := testBodyCompress(t, bodyHTML)

	type want struct {
		contentEncoding string
		body            string
	}

	tests := []struct {
		name           string
		statusCode     int
		acceptEncoding string
		contectType    string
		body           string
		want           want
	}{
		{
			name:           "compressed_json",
			statusCode:     http.StatusOK,
			acceptEncoding: "gzip",
			contectType:    "application/json",
			body:           bodyJSON,
			want: want{
				contentEncoding: "gzip",
				body:            bodyJSONgz,
			},
		},
		{
			name:           "not_compressed_json_without_header",
			statusCode:     http.StatusOK,
			acceptEncoding: "",
			contectType:    "application/json",
			body:           bodyJSON,
			want: want{
				contentEncoding: "",
				body:            bodyJSON,
			},
		},
		{
			name:           "not_compressed_json_status_code_not_ok",
			statusCode:     http.StatusBadRequest,
			acceptEncoding: "gzip",
			contectType:    "application/json",
			body:           bodyJSON,
			want: want{
				contentEncoding: "",
				body:            bodyJSON,
			},
		},
		{
			name:           "compressed_html",
			statusCode:     http.StatusOK,
			acceptEncoding: "gzip",
			contectType:    "text/html",
			body:           bodyHTML,
			want: want{
				contentEncoding: "gzip",
				body:            bodyHTMLgz,
			},
		},
		{
			name:           "not_compressed_html_without_header",
			statusCode:     http.StatusOK,
			acceptEncoding: "",
			contectType:    "text/html",
			body:           bodyHTML,
			want: want{
				contentEncoding: "",
				body:            bodyHTML,
			},
		},
		{
			name:           "not_compressed_html_status_code_not_ok",
			statusCode:     http.StatusBadRequest,
			acceptEncoding: "gzip",
			contectType:    "text/html",
			body:           bodyHTML,
			want: want{
				contentEncoding: "",
				body:            bodyHTML,
			},
		},
		{
			name:           "not_compressed_not_html_or_json",
			statusCode:     http.StatusOK,
			acceptEncoding: "gzip",
			contectType:    "text/plain",
			body:           "some text",
			want: want{
				contentEncoding: "",
				body:            "some text",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.acceptEncoding != "" {
				r.Header.Set("Accept-Encoding", test.acceptEncoding)
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", test.contectType)
				w.WriteHeader(test.statusCode)
				io.WriteString(w, test.body)
			})

			w := httptest.NewRecorder()
			handler := GzipWrapper(next)
			handler.ServeHTTP(w, r)
			res := w.Result()
			defer res.Body.Close()

			encoding := res.Header.Get("Content-Encoding")
			assert.Equal(t, test.want.contentEncoding, encoding, "Response Content-Encoding")
			resBody, err := io.ReadAll(res.Body)
			require.Nil(t, err, "Read response body")
			body := strings.TrimSuffix(string(resBody), "\n")
			assert.Equal(t, test.want.body, body, "Response body")
		})
	}
}
