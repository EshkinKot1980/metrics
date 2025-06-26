package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashHeaderValidate(t *testing.T) {
	secret := "secret"
	body := []byte(`{"data": "some data"}`)

	type want struct {
		code int
		body string
	}

	tests := []struct {
		name    string
		secret  string
		reqHash string
		reqBody []byte
		want    want
	}{
		{
			name:    "positive",
			secret:  secret,
			reqHash: hashHexString(body, secret),
			reqBody: body,
			want: want{
				code: http.StatusOK,
				body: "",
			},
		},
		{
			name:    "positive_without_secret",
			reqBody: body,
			want: want{
				code: http.StatusOK,
				body: "",
			},
		},
		{
			name:    "negative_without_hash",
			secret:  secret,
			reqBody: body,
			want: want{
				code: http.StatusBadRequest,
				body: "empty hash header",
			},
		},
		{
			name:    "negative_invalid_hash",
			secret:  secret,
			reqHash: "invalid",
			reqBody: body,
			want: want{
				code: http.StatusBadRequest,
				body: "invalid hash",
			},
		},
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(test.reqBody))
			if test.secret != "" && test.reqHash != "" {
				r.Header.Set("HashSHA256", test.reqHash)
			}

			w := httptest.NewRecorder()
			mw := NewHashHeader(test.secret)
			handler := mw.Validate(nextHandler)
			handler.ServeHTTP(w, r)
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

func TestHashHeaderSign(t *testing.T) {
	secret := "secret"
	body := []byte(`{"data": "some data"}`)

	tests := []struct {
		name     string
		secret   string
		respBody []byte
		wantHash string
	}{
		{
			name:     "signed",
			secret:   secret,
			respBody: body,
			wantHash: hashHexString(body, secret),
		},
		{
			name:     "unsigned_without_secret",
			respBody: body,
			wantHash: "",
		},
		{
			name:     "unsigned_empty_body",
			secret:   secret,
			respBody: []byte{},
			wantHash: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				if len(test.respBody) > 0 {
					_, err := w.Write(test.respBody)
					assert.Nil(t, err, "Response body write")
				}
			})

			w := httptest.NewRecorder()
			mw := NewHashHeader(test.secret)
			handler := mw.Sign(next)
			handler.ServeHTTP(w, r)

			assert.Equal(t, test.wantHash, w.Header().Get("HashSHA256"), "Response hash header")
		})
	}
}
