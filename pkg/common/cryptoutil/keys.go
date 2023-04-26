package cryptoutil

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	rsaPrivateKeyType = "RSA PRIVATE KEY"
	ecPrivateKeyType  = "EC PRIVATE KEY"
)

// KeyType represents the types of keys.
type KeyType int

const (
	KeyTypeUnset KeyType = iota
	ECP256
	ECP384
	RSA2048
	RSA4096
)

// GenerateSigner generates a new key for the given key type.
func GenerateSigner(keyType KeyType) (crypto.Signer, error) {
	switch keyType {
	case ECP256:
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case ECP384:
		return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case RSA2048:
		return rsa.GenerateKey(rand.Reader, 2048)
	case RSA4096:
		return rsa.GenerateKey(rand.Reader, 4096)
	}
	return nil, fmt.Errorf("unknown key type %q", keyType)
}

func (keyType KeyType) String() string {
	switch keyType {
	case KeyTypeUnset:
		return "UNSET"
	case ECP256:
		return "ec-p256"
	case ECP384:
		return "ec-p384"
	case RSA2048:
		return "rsa-2048"
	case RSA4096:
		return "rsa-4096"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(keyType))
	}
}

// LoadPrivateKey loads a private key from file in PEM format.
// The key can be either an RSA or EC private key.
func LoadPrivateKey(path string) (crypto.PrivateKey, error) {
	keyFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed reading private key: %w", err)
	}

	header := strings.Split(string(keyFile), "\n")[0]
	if strings.Contains(header, rsaPrivateKeyType) {
		key, err := ParseRSAPrivateKeyPEM(keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed parsing private key: %w", err)
		}
		return key, nil
	}

	if strings.Contains(header, ecPrivateKeyType) {
		key, err := ParseECPrivateKeyPEM(keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed parsing private key: %w", err)
		}
		return key, nil
	}

	return nil, errors.New("private key format not supported")
}

// LoadRSAPrivateKey loads an RSA private key from a file.
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

// LoadECPrivateKey loads an EC private key from a file.
func LoadECPrivateKey(path string) (crypto.PrivateKey, error) {
	keyFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed reading private key: %w", err)
	}

	key, err := ParseECPrivateKeyPEM(keyFile)
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

// ParseRSAPrivateKeyPEM parses an RSA private key in PEM format.
func ParseRSAPrivateKeyPEM(pemBlocks []byte) (interface{}, error) {
	block, _ := pem.Decode(pemBlocks)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	return ParseRSAPrivateKey(block.Bytes)
}

// EncodeRSAPrivateKey encodes an RSA private key in PEM format.
func EncodeRSAPrivateKey(privateKey *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  rsaPrivateKeyType,
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
}

// ParseECPrivateKey parses an EC private key in PKCS #1, ASN.1 DER form.
func ParseECPrivateKey(derBytes []byte) (crypto.PrivateKey, error) {
	key, err := x509.ParseECPrivateKey(derBytes)
	if err != nil {
		return nil, fmt.Errorf("failed parsing private key: %w", err)
	}

	return key, nil
}

// ParseECPrivateKeyPEM parses an RSA private key in PEM format.
func ParseECPrivateKeyPEM(pemBlocks []byte) (interface{}, error) {
	block, _ := pem.Decode(pemBlocks)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	return ParseECPrivateKey(block.Bytes)
}

// EncodeECPrivateKey encodes an RSA private key in PEM format.
func EncodeECPrivateKey(privateKey *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  ecPrivateKeyType,
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
}
