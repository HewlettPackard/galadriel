package cryptoutil_test

import (
	"crypto/rsa"
	"os"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/stretchr/testify/require"
)

const (
	certPEM   = "testdata/cert.pem"
	rsaKeyPEM = "testdata/rsa-key.pem"
)

func TestParsePrivateKeyPEM(t *testing.T) {
	// not a private key
	_, err := cryptoutil.ParseRSAPrivateKeyPEM(readFile(t, certPEM))
	require.Error(t, err)

	// success with RSA
	key, err := cryptoutil.ParseRSAPrivateKeyPEM(readFile(t, rsaKeyPEM))

	require.NoError(t, err)
	require.NotNil(t, key)
	_, ok := key.(*rsa.PrivateKey)
	require.True(t, ok)
}

func TestLoadRSAPrivateKey(t *testing.T) {
	// not a private key
	_, err := cryptoutil.LoadRSAPrivateKey(certPEM)
	require.Error(t, err)

	// success with RSA
	key, err := cryptoutil.LoadRSAPrivateKey(rsaKeyPEM)
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
	_, err := cryptoutil.LoadCertificate(rsaKeyPEM)
	require.Error(t, err)

	// success
	cert, err := cryptoutil.LoadCertificate(certPEM)
	require.NoError(t, err)
	require.NotNil(t, cert)
}

func TestParseCertificate(t *testing.T) {
	// not a certificate
	_, err := cryptoutil.ParseCertificate(readFile(t, rsaKeyPEM))
	require.Error(t, err)

	// success with one certificate
	cert, err := cryptoutil.ParseCertificate(readFile(t, certPEM))
	require.NoError(t, err)
	require.NotNil(t, cert)
}

func TestEncodeCertificates(t *testing.T) {
	cert, err := cryptoutil.LoadCertificate(certPEM)
	require.NoError(t, err)
	expCertPem, err := os.ReadFile(certPEM)
	require.NoError(t, err)
	require.Equal(t, expCertPem, cryptoutil.EncodeCertificate(cert))
}
