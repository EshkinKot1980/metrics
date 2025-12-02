// Модуль middleware реализует промежуточный слой обработки HTTP запроса.
package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/EshkinKot1980/metrics/internal/common/utils"
)

type RSA struct {
	privateKey *rsa.PrivateKey
}

func NewRSA(key *rsa.PrivateKey) *RSA {
	return &RSA{privateKey: key}
}

func (mw *RSA) Decrypt(next http.Handler) http.Handler {
	if mw.privateKey == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		data, err := utils.DecryptWithPrivateKey(body, mw.privateKey)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(data))
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
