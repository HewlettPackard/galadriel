package certtest

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"os"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/jmhodges/clock"
	"github.com/stretchr/testify/require"
)

const (
	keyPerm = 0600
	crtPerm = 0644
	oneHour = time.Hour
)

// CreateTestSelfSignedCACertificate creates a self-signed CA certificate for testing purposes.
func CreateTestSelfSignedCACertificate(t *testing.T, clk clock.Clock) (*x509.Certificate, crypto.PrivateKey) {
	name := pkix.Name{CommonName: "root-ca"}

	template, err := cryptoutil.CreateRootCATemplate(clk, name, oneHour)
	require.NoError(t, err)

	caCert, privateKey, err := cryptoutil.SelfSignX509(template)
	require.NoError(t, err)

	return caCert, privateKey
}

// CreateTestIntermediateCACertificate creates an intermediate CA certificate for testing purposes.
func CreateTestIntermediateCACertificate(t *testing.T, clk clock.Clock, parent *x509.Certificate, parentKey crypto.PrivateKey, cn string) (*x509.Certificate, crypto.PrivateKey) {
	name := pkix.Name{CommonName: cn}

	signer, err := cryptoutil.GenerateSigner(cryptoutil.DefaultKeyType)
	require.NoError(t, err)

	template, err := cryptoutil.CreateCATemplate(clk, signer.Public(), name, oneHour)
	require.NoError(t, err)

	caCert, err := cryptoutil.SignX509(template, parent, parentKey)
	require.NoError(t, err)

	return caCert, signer
}

// CreateTestCACertificates creates a self-signed CA and an intermediate CA for testing purposes.
// It returns the path to the temporary directory where the certificates are stored.
func CreateTestCACertificates(t *testing.T, clk clock.Clock) string {
	tempDir, err := os.MkdirTemp("", "galadriel-test")
	require.NoError(t, err)

	// create a Self-signed CA
	rootCA, rootKey := CreateTestSelfSignedCACertificate(t, clk)

	pemCert := cryptoutil.EncodeCertificate(rootCA)
	rsaKey, ok := rootKey.(*rsa.PrivateKey)
	require.True(t, ok)

	// write the Root CA certificate to disk
	err = os.WriteFile(tempDir+"/root-ca.crt", pemCert, crtPerm)
	require.NoError(t, err)
	err = os.WriteFile(tempDir+"/root-ca.key", cryptoutil.EncodeRSAPrivateKey(rsaKey), keyPerm)
	require.NoError(t, err)

	// create intermediate CA signed by the Self-signed CA
	intermediateCA, intermediateKey := CreateTestIntermediateCACertificate(t, clk, rootCA, rootKey, "intermediate-ca")

	pemCert = cryptoutil.EncodeCertificate(intermediateCA)
	rsaKey, ok = intermediateKey.(*rsa.PrivateKey)
	require.True(t, ok)

	// write the intermediate CA certificate to disk
	err = os.WriteFile(tempDir+"/intermediate-ca.crt", pemCert, crtPerm)
	require.NoError(t, err)
	err = os.WriteFile(tempDir+"/intermediate-ca.key", cryptoutil.EncodeRSAPrivateKey(rsaKey), keyPerm)
	require.NoError(t, err)

	// create intermediate-CA-2 signed by the intermediate CA-1
	intermediateCA2, intermediateKey2 := CreateTestIntermediateCACertificate(t, clk, intermediateCA, intermediateKey, "intermediate-ca-2")

	pemCert = cryptoutil.EncodeCertificate(intermediateCA2)
	rsaKey, ok = intermediateKey2.(*rsa.PrivateKey)
	require.True(t, ok)

	// write the intermediate CA certificate to disk
	err = os.WriteFile(tempDir+"/intermediate-ca-2.crt", pemCert, crtPerm)
	require.NoError(t, err)
	err = os.WriteFile(tempDir+"/intermediate-ca-2.key", cryptoutil.EncodeRSAPrivateKey(rsaKey), keyPerm)
	require.NoError(t, err)

	// create a other Self-signed CA
	otherCA, otherKey := CreateTestSelfSignedCACertificate(t, clk)

	pemCert = cryptoutil.EncodeCertificate(otherCA)
	rsaKey, ok = otherKey.(*rsa.PrivateKey)
	require.True(t, ok)

	// write the Root CA certificate to disk
	err = os.WriteFile(tempDir+"/other-ca.crt", pemCert, crtPerm)
	require.NoError(t, err)
	err = os.WriteFile(tempDir+"/other-ca.key", cryptoutil.EncodeRSAPrivateKey(rsaKey), keyPerm)
	require.NoError(t, err)

	return tempDir
}
