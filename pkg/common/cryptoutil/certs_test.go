package cryptoutil

import (
	"crypto/rsa"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePrivateKeyPEM(t *testing.T) {
	// not a private key
	_, err := ParseRSAPrivateKeyPEM(readFile(t, "testdata/cert.pem"))
	require.Error(t, err)

	// success with RSA
	key, err := ParseRSAPrivateKeyPEM(readFile(t, "testdata/rsa-key.pem"))

	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok := key.(*rsa.PrivateKey)
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
}

func readFile(t *testing.T, path string) []byte {
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}

func TestLoadCertificate(t *testing.T) {
	// not a certificate
	_, err := LoadCertificate("testdata/rsa-key.pem")
	require.Error(t, err)

	// success
	cert, err := LoadCertificate("testdata/cert.pem")
	require.NoError(t, err)
	require.NotNil(t, cert)
}

func TestParseCertificate(t *testing.T) {
	// not a certificate
	_, err := ParseCertificate(readFile(t, "testdata/rsa-key.pem"))
	require.Error(t, err)

	// success with one certificate
	cert, err := ParseCertificate(readFile(t, "testdata/cert.pem"))
	require.NoError(t, err)
	require.NotNil(t, cert)
}

func TestEncodeCertificates(t *testing.T) {
	cert, err := LoadCertificate("testdata/cert.pem")
	require.NoError(t, err)
	expCertPem, err := os.ReadFile("testdata/cert.pem")
	require.NoError(t, err)
	require.Equal(t, expCertPem, EncodeCertificate(cert))
}
