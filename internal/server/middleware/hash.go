package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

type HashHeader struct {
	secret string
}

func NewHashHeader(secretKey string) *HashHeader {
	return &HashHeader{secret: secretKey}
}

func (h *HashHeader) Validate(next http.Handler) http.Handler {
	if h.secret == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		requestHash := r.Header.Get("HashSHA256")
		if requestHash == "" {
			http.Error(w, "empty hash header", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		hash := hashHexString(body, h.secret)
		if hash != requestHash {
			http.Error(w, "invalid hash", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func (h *HashHeader) Sign(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		writer := w
		if h.secret != "" {
			writer = newSignWriter(w, h.secret)
		}

		next.ServeHTTP(writer, r)
	}

	return http.HandlerFunc(fn)
}

type signWriter struct {
	secret string
	code   int
	w      http.ResponseWriter
}

func newSignWriter(w http.ResponseWriter, secret string) *signWriter {
	return &signWriter{w: w, secret: secret}
}

func (s *signWriter) Write(p []byte) (int, error) {
	hash := hashHexString(p, s.secret)
	s.w.Header().Set("HashSHA256", hash)
	s.w.WriteHeader(s.code)
	return s.w.Write(p)
}

func (s *signWriter) Header() http.Header {
	return s.w.Header()
}

func (s *signWriter) WriteHeader(statusCode int) {
	contentType := s.w.Header().Get("Content-Type")
	if contentType == "" {
		s.w.WriteHeader(statusCode)
	} else {
		s.code = statusCode
	}
}

func hashHexString(body []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
