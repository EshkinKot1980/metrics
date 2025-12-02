package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRSA(t *testing.T) {
	data := []byte("Some test data")

	dir, err := os.Getwd()
	require.Nil(t, err, "Utils directorytils")
	publicKeyPath := filepath.Join(dir, "testdata/public-key.pem")
	privateKeyPath := filepath.Join(dir, "testdata/private-key.pem")

	publicKey, err := LoadPublicKey(publicKeyPath)
	require.Nil(t, err, "Load public key from file")

	encryptedData, err := EncryptWithPublicKey(data, publicKey)
	require.Nil(t, err, "Encrypt data")

	privateKey, err := LoadPrivateKey(privateKeyPath)
	require.Nil(t, err, "Load private key from file")

	decryptedData, err := DecryptWithPrivateKey(encryptedData, privateKey)
	require.Nil(t, err, "Decrypt data")

	require.Equal(t, data, decryptedData, "Check data equal")
}
