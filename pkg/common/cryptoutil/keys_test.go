package cryptoutil

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"github.com/stretchr/testify/require"
	"testing"
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
	_, err := ParseRSAPrivateKeyPEM(readFile(t, "testdata/cert.pem"))
	require.Error(t, err)

	// success with RSA
	key, err := ParseRSAPrivateKeyPEM(readFile(t, "testdata/rsa-key.pem"))
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok := key.(*rsa.PrivateKey)
	require.True(t, ok)

	// failure with EC
	key, err = ParseRSAPrivateKeyPEM(readFile(t, "testdata/ec-key.pem"))
	require.Error(t, err)
	require.Nil(t, key)
}

func TestParseECPrivateKeyPEM(t *testing.T) {
	// not a private key
	_, err := ParseECPrivateKeyPEM(readFile(t, "testdata/cert.pem"))
	require.Error(t, err)

	// success with EC
	key, err := ParseECPrivateKeyPEM(readFile(t, "testdata/ec-key.pem"))
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok := key.(*ecdsa.PrivateKey)
	require.True(t, ok)

	// failure with RSA
	key, err = ParseECPrivateKeyPEM(readFile(t, "testdata/rsa-key.pem"))
	require.Error(t, err)
	require.Nil(t, key)
}

func TestLoadPrivateKey(t *testing.T) {
	// not a private key
	_, err := LoadPrivateKey("testdata/cert.pem")
	require.Error(t, err)

	// success with RSA
	key, err := LoadPrivateKey("testdata/rsa-key.pem")
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok := key.(*rsa.PrivateKey)
	require.True(t, ok)

	// success with EC
	key, err = LoadPrivateKey("testdata/ec-key.pem")
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok = key.(*ecdsa.PrivateKey)
	require.True(t, ok)
}

func TestLoadRSAPrivateKey(t *testing.T) {
	// not a private key
	_, err := LoadRSAPrivateKey("testdata/cert.pem")
	require.Error(t, err)

	// success with RSA
	key, err := LoadRSAPrivateKey("testdata/rsa-key.pem")
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok := key.(*rsa.PrivateKey)
	require.True(t, ok)

	// failure with EC
	key, err = LoadRSAPrivateKey("testdata/rsa-ec.pem")
	require.Error(t, err)
	require.Nil(t, key)
}

func TestLoadECPrivateKey(t *testing.T) {
	// not a private key
	_, err := LoadECPrivateKey("testdata/cert.pem")
	require.Error(t, err)

	// success with EC
	key, err := LoadECPrivateKey("testdata/ec-key.pem")
	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok := key.(*ecdsa.PrivateKey)
	require.True(t, ok)

	// failure with RSA
	key, err = LoadECPrivateKey("testdata/rsa-key.pem")
	require.Error(t, err)
	require.Nil(t, key)
}
