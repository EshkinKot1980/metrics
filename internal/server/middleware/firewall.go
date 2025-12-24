// Модуль middleware реализует промежуточный слой обработки HTTP запроса.
package middleware

import (
	"net"
	"net/http"
)

type Firewall struct {
	trustedNet *net.IPNet
}

func NewFirewall(trusted *net.IPNet) *Firewall {
	return &Firewall{trustedNet: trusted}
}

func (mw *Firewall) Filter(next http.Handler) http.Handler {
	if mw.trustedNet == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		clientIP := net.ParseIP(r.Header.Get("X-Real-IP"))

		if !mw.trustedNet.Contains(clientIP) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
