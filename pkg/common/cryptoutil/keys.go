package cryptoutil

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

func CreateRSAKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed generating private key: %w", err)
	}

	return privateKey, nil
}

func LoadRSAPrivateKey(path string) (crypto.PrivateKey, error) {
	keyFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed reading private key: %w", err)
	}

	key, err := ParseRSAPrivateKeyPEM(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed parsing private key: %w", err)
	}

	return key, nil
}

// ParseRSAPrivateKey parses an RSA private key in PKCS #1, ASN.1 DER form.
func ParseRSAPrivateKey(derBytes []byte) (crypto.PrivateKey, error) {
	key, err := x509.ParsePKCS1PrivateKey(derBytes)
	if err != nil {
		return nil, fmt.Errorf("failed parsing private key: %w", err)
	}

	return key, nil
}

func ParseRSAPrivateKeyPEM(pemBlocks []byte) (interface{}, error) {
	block, _ := pem.Decode(pemBlocks)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	return ParseRSAPrivateKey(block.Bytes)
}

func EncodeRSAPrivateKey(privateKey *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
}
