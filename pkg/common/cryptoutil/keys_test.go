package cryptoutil

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	certPath      = "testdata/cert.pem"
	certChainPath = "testdata/cert-chain.pem"
	rsaKeyPath    = "testdata/rsa-key.pem"
	ecKeyPath     = "testdata/ec-key.pem"
)

func TestGenerateSigner(t *testing.T) {
	// success with RSA
	key, err := GenerateSigner(RSA2048)
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok := key.(*rsa.PrivateKey)
	require.True(t, ok)

	key, err = GenerateSigner(RSA4096)
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok = key.(*rsa.PrivateKey)
	require.True(t, ok)

	// success with EC
	key, err = GenerateSigner(ECP256)
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok = key.(*ecdsa.PrivateKey)
	require.True(t, ok)

	key, err = GenerateSigner(ECP384)
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok = key.(*ecdsa.PrivateKey)
	require.True(t, ok)
}

func TestParseRSAPrivateKeyPEM(t *testing.T) {
	// not a private key
	_, err := ParseRSAPrivateKeyPEM(readFile(t, certPath))
	require.Error(t, err)

	// success with RSA
	key, err := ParseRSAPrivateKeyPEM(readFile(t, rsaKeyPath))
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok := key.(*rsa.PrivateKey)
	require.True(t, ok)

	// failure with EC
	key, err = ParseRSAPrivateKeyPEM(readFile(t, ecKeyPath))
	require.Error(t, err)
	require.Nil(t, key)
}

func TestParseECPrivateKeyPEM(t *testing.T) {
	// not a private key
	_, err := ParseECPrivateKeyPEM(readFile(t, certPath))
	require.Error(t, err)

	// success with EC
	key, err := ParseECPrivateKeyPEM(readFile(t, ecKeyPath))
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok := key.(*ecdsa.PrivateKey)
	require.True(t, ok)

	// failure with RSA
	key, err = ParseECPrivateKeyPEM(readFile(t, rsaKeyPath))
	require.Error(t, err)
	require.Nil(t, key)
}

func TestLoadPrivateKey(t *testing.T) {
	// not a private key
	_, err := LoadPrivateKey(certPath)
	require.Error(t, err)

	// success with RSA
	key, err := LoadPrivateKey(rsaKeyPath)
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok := key.(*rsa.PrivateKey)
	require.True(t, ok)

	// success with EC
	key, err = LoadPrivateKey(ecKeyPath)
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok = key.(*ecdsa.PrivateKey)
	require.True(t, ok)
}

func TestLoadRSAPrivateKey(t *testing.T) {
	// not a private key
	_, err := LoadRSAPrivateKey(certPath)
	require.Error(t, err)

	// success with RSA
	key, err := LoadRSAPrivateKey(rsaKeyPath)
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok := key.(*rsa.PrivateKey)
	require.True(t, ok)

	// failure with EC
	key, err = LoadRSAPrivateKey(ecKeyPath)
	require.Error(t, err)
	require.Nil(t, key)
}

func TestLoadECPrivateKey(t *testing.T) {
	// not a private key
	_, err := LoadECPrivateKey(certPath)
	require.Error(t, err)

	// success with EC
	key, err := LoadECPrivateKey(ecKeyPath)
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok := key.(*ecdsa.PrivateKey)
	require.True(t, ok)

	// failure with RSA
	key, err = LoadECPrivateKey(rsaPrivateKeyType)
	require.Error(t, err)
	require.Nil(t, key)
}

func TestEncodeRSAPrivateKey(t *testing.T) {
	key, err := LoadRSAPrivateKey(rsaKeyPath)
	require.NoError(t, err)
	require.NotNil(t, key)

	pem := EncodeRSAPrivateKey(key.(*rsa.PrivateKey))
	require.NoError(t, err)
	require.NotNil(t, pem)
}

func TestEncodeECPrivateKey(t *testing.T) {
	key, err := LoadECPrivateKey(ecKeyPath)
	require.NoError(t, err)
	require.NotNil(t, key)

	pem, err := EncodeECPrivateKey(key.(*ecdsa.PrivateKey))
	require.NoError(t, err)
	require.NoError(t, err)
	require.NotNil(t, pem)
}

func TestVerifyCertificatePrivateKeyRSA(t *testing.T) {
	cert, key := createCert(t, RSA4096)

	err := VerifyCertificatePrivateKey(cert, key)
	require.NoError(t, err)

	// failure with wrong key
	key, err = GenerateSigner(RSA4096)
	require.NoError(t, err)
	require.NotNil(t, key)
	err = VerifyCertificatePrivateKey(cert, key)
	require.Error(t, err)
	assert.Equal(t, "certificate public key does not match private key", err.Error())
}

func TestVerifyCertificatePrivateKeyEC(t *testing.T) {
	cert, key := createCert(t, ECP384)

	err := VerifyCertificatePrivateKey(cert, key)
	require.NoError(t, err)

	// failure with wrong key
	key, err = GenerateSigner(ECP256)
	require.NoError(t, err)
	require.NotNil(t, key)
	err = VerifyCertificatePrivateKey(cert, key)
	require.Error(t, err)
	assert.Equal(t, "certificate public key does not match private key", err.Error())
}
