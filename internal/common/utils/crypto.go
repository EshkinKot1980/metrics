package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

var ErrInvalidKey = errors.New("invalid key")

// LoadPublicKey загружает публичный ключ из файла в формате pem.
func LoadPublicKey(fileName string) (*rsa.PublicKey, error) {
	if fileName == "" {
		return nil, nil
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to pem.Decode public key: %w", ErrInvalidKey)
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid public key: %w", ErrInvalidKey)
	}

	return publicKey, nil
}

// LoadPrivateKey загружает приватный ключ из файла в формате pem.
func LoadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	if filename == "" {
		return nil, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to pem.Decode private key: %w", ErrInvalidKey)
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	privateKey, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid private key: %w", ErrInvalidKey)
	}

	return privateKey, nil
}

// EncryptWithPublicKey шифрует данные с помощью публичного ключа.
func EncryptWithPublicKey(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	if publicKey == nil {
		return data, nil
	}

	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	encryptedData := gcm.Seal(nonce, nonce, data, nil)

	encryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt AES key with RSA: %w", err)
	}

	keySize := len(encryptedKey)
	result := make([]byte, 4+keySize+len(encryptedData))

	binary.BigEndian.PutUint32(result[0:4], uint32(keySize))

	copy(result[4:4+keySize], encryptedKey)
	copy(result[4+keySize:], encryptedData)

	return result, nil
}

// DecryptWithPrivateKey расшифровывает данные с помощью приватного ключа
func DecryptWithPrivateKey(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return data, nil
	}

	if len(data) < 4 {
		return nil, fmt.Errorf("invalid encrypted data: too short for key size header")
	}

	keySize := int(binary.BigEndian.Uint32(data[0:4]))

	if len(data) < 4+keySize {
		return nil, fmt.Errorf("encrypted data too short, need %d bytes for key, have %d", 4+keySize, len(data))
	}

	encryptedKey := data[4 : 4+keySize]
	encryptedData := data[4+keySize:]

	aesKey, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()

	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("invalid encrypted data: missing nonce")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	decryptedData, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return decryptedData, nil
}

// GenerateKeyPair создает пару ключей rsa для тестов.
func GenerateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privkey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dgenerate key pair: %w", err)
	}
	return privkey, &privkey.PublicKey, nil
}
