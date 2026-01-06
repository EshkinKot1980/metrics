package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirewall_Filter(t *testing.T) {
	_, trustedNet, err := net.ParseCIDR("172.17.0.0/16")
	require.Nil(t, err, "Parsing trusted network")

	tests := []struct {
		name     string
		ip       string
		net      *net.IPNet
		wantCode int
	}{
		{
			name:     "success_without_filter",
			ip:       "",
			net:      nil,
			wantCode: http.StatusOK,
		},
		{
			name:     "success_with_filter",
			ip:       "172.17.1.2",
			net:      trustedNet,
			wantCode: http.StatusOK,
		},
		{
			name:     "negative_without_header",
			ip:       "",
			net:      trustedNet,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "negative_invalid_ip",
			ip:       "InvalidIP",
			net:      trustedNet,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "negative_untrusted_network",
			ip:       "172.16.1.1",
			net:      trustedNet,
			wantCode: http.StatusUnauthorized,
		},
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.ip != "" {
				r.Header.Set("X-Real-IP", test.ip)
			}

			w := httptest.NewRecorder()
			mw := NewFirewall(test.net)
			handler := mw.Filter(nextHandler)
			handler.ServeHTTP(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.wantCode, res.StatusCode, "Response status code")

		})
	}
}
