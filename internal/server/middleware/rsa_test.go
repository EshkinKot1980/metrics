package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EshkinKot1980/metrics/internal/common/utils"
)

func TestRSA_Decrypt(t *testing.T) {
	body := []byte(`{"data": "some data"}`)

	priv1, pub1, err := utils.GenerateKeyPair()
	require.Nil(t, err, "Generate rsa key pair")

	priv2, _, err := utils.GenerateKeyPair()
	require.Nil(t, err, "Generate rsa key pair")

	tests := []struct {
		name     string
		pub      *rsa.PublicKey
		priv     *rsa.PrivateKey
		wantCode int
	}{
		{
			name:     "succes_not_encrypted",
			wantCode: http.StatusOK,
		},
		{
			name:     "succes_encrypted",
			pub:      pub1,
			priv:     priv1,
			wantCode: http.StatusOK,
		},
		{
			name:     "succes_encrypted",
			pub:      pub1,
			priv:     priv2,
			wantCode: http.StatusUnauthorized,
		},
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, err := io.ReadAll(r.Body)
		require.Nil(t, err, "Reading body")
		assert.Equal(t, body, reqBody, "Decrypted body")
		w.WriteHeader(http.StatusOK)
	})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var reqBody *bytes.Buffer

			if test.pub == nil {
				reqBody = bytes.NewBuffer(body)
			} else {
				encrData, err := utils.EncryptWithPublicKey(body, test.pub)
				require.Nil(t, err, "Request Body encrypting")
				reqBody = bytes.NewBuffer(encrData)
			}

			r := httptest.NewRequest(http.MethodPost, "/", reqBody)

			w := httptest.NewRecorder()
			mw := NewRSA(test.priv)
			handler := mw.Decrypt(nextHandler)
			handler.ServeHTTP(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.wantCode, res.StatusCode, "Response status code")
		})
	}
}
