package middleware

import (
	"net/http"

	"github.com/EshkinKot1980/metrics/internal/server"
)

func ValidateMetric(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		m := server.ParsePath(r)
		if err := m.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
