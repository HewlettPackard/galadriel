package integrity

import (
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/test/certtest"
	"github.com/jmhodges/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var clk = clock.NewFake()

const (
	rootCACert = "/root-ca.crt"
	rootCAKey  = "/root-ca.key"
)

func TestDiskSignerAndVerifierUsingRootCA(t *testing.T) {
	var signer Signer

	caTemplate, err := cryptoutil.CreateCATemplate(clk, nil, pkix.Name{CommonName: "test-ca"}, 1*time.Hour)
	require.NoError(t, err)

	rootCert, privateKey, err := cryptoutil.SelfSignX509(caTemplate)
	require.NoError(t, err)
	signerKey, ok := privateKey.(crypto.Signer)
	require.True(t, ok)

	// create a DiskSigner using the root key and the root certificate
	signer = &DiskSigner{
		caPrivateKey: signerKey,
		caCert:       rootCert,
		clk:          clk,
	}

	createPayloadSignVerify(t, signer, rootCert)

}

func TestDiskSignerAndVerifierUsingIntermediateCA(t *testing.T) {
	var signer Signer

	rootCaTemplate, err := cryptoutil.CreateCATemplate(clk, nil, pkix.Name{CommonName: "test-ca"}, 1*time.Hour)
	require.NoError(t, err)

	rootCert, rootPrivateKey, err := cryptoutil.SelfSignX509(rootCaTemplate)
	require.NoError(t, err)

	// create key pair for intermediate CA
	intermediateKey, err := cryptoutil.GenerateSigner(cryptoutil.DefaultKeyType)
	require.NoError(t, err)
	intermediateCaTemplate, err := cryptoutil.CreateCATemplate(clk, intermediateKey.Public(), pkix.Name{CommonName: "test-intermediate-ca"}, 1*time.Hour)
	require.NoError(t, err)
	intermediateCert, err := cryptoutil.SignX509(intermediateCaTemplate, rootCaTemplate, rootPrivateKey)
	require.NoError(t, err)

	// create a DiskSigner using intermediateKey and the intermediate certificate
	signer = &DiskSigner{
		caPrivateKey: intermediateKey,
		caCert:       intermediateCert,
		clk:          clk,
	}
	createPayloadSignVerify(t, signer, rootCert)
}

// TestNewDiskSignerConfig tests for a new config with string parameters for the key and cert paths
func TestNewDisSignerConfig(t *testing.T) {

	// create selfsigned root CA
	tmpDir := certtest.CreateTestCACertificates(t, clk)
	assert.NotNil(t, tmpDir)

	tmpCertPath := tmpDir + rootCACert
	tmpKeyPath := tmpDir + rootCAKey

	newConf := NewDiskSignerConfig(tmpKeyPath, tmpCertPath)
	assert.NotNil(t, newConf)
	assert.Equal(t, tmpCertPath, newConf.CACertPath)
	assert.Equal(t, tmpKeyPath, newConf.CAPrivateKeyPath)

}

// TestNewDisVerifierConfig test for a new config  string parameters for the trust bundle path
func TestNewDisVerifierConfig(t *testing.T) {

	// create selfsigned root CA
	tmpDir := certtest.CreateTestCACertificates(t, clk)
	assert.NotNil(t, tmpDir)

	tmpCertPath := tmpDir + rootCACert

	newConf := NewDiskVerifierConfig(tmpCertPath)
	assert.NotNil(t, newConf)
	assert.Equal(t, tmpCertPath, newConf.TrustBundlePath)
}

// TestNewDiskSigner tests the creation of a NewDiskSigner with DiskSignerConfig
func TestNewDiskSigner(t *testing.T) {

	// create selfsigned root CA
	tmpDir := certtest.CreateTestCACertificates(t, clk)
	assert.NotNil(t, tmpDir)

	tmpCertPath := tmpDir + rootCACert
	tmpKeyPath := tmpDir + rootCAKey

	newConf := NewDiskSignerConfig(tmpKeyPath, tmpCertPath)
	assert.NotNil(t, newConf)

	newSigner, err := NewDiskSigner(newConf)
	assert.NotNil(t, newSigner)
	assert.NoError(t, err)
}

// TestNewDiskSignerWithInvalidConfig tests for the error when paths in config are invalid
func TestNewDiskSignerWithInvalidConfig(t *testing.T) {

	newConf := NewDiskSignerConfig("invalidpath", "invalidpath")
	assert.NotNil(t, newConf)

	// create a new signer with invalid config
	newConf.CACertPath = "invalid"
	newSigner, err := NewDiskSigner(newConf)
	assert.Nil(t, newSigner)
	assert.Error(t, err)
}

// TestNewDiskVerifier tests the creation of a NewDiskVerifier with DiskVerifierConfig
func TestNewDiskVerifier(t *testing.T) {

	// create selfsigned root CA
	tmpDir := certtest.CreateTestCACertificates(t, clk)
	assert.NotNil(t, tmpDir)

	tmpCertPath := tmpDir + rootCACert

	newConf := NewDiskVerifierConfig(tmpCertPath)
	assert.NotNil(t, newConf)

	newVerifier, err := NewDiskVerifier(newConf)
	assert.NotNil(t, newVerifier)
	assert.NoError(t, err)
}

func createPayloadSignVerify(t *testing.T, signer Signer, rootCert *x509.Certificate) {
	var verifier Verifier
	// create a test payload
	payload := []byte("test payload")
	otherPayload := []byte("other payload")

	// sign the payload using the caPrivateKey
	signature, signingCert, err := signer.Sign(payload)
	require.NoError(t, err)
	require.NotNil(t, signingCert)

	// sign the payload using the caPrivateKey
	otherSignature, otherCert, err := signer.Sign(otherPayload)
	require.NoError(t, err)
	require.NotNil(t, otherCert)

	// create a verifier using the root certificate in the trust bundle
	verifier = &DiskVerifier{
		trustBundle: []*x509.Certificate{rootCert},
		clk:         clk,
	}

	// verify the signature using the verifier
	err = verifier.Verify(payload, signature, signingCert)
	require.NoError(t, err)

	err = verifier.Verify(otherPayload, otherSignature, otherCert)
	require.NoError(t, err)

	err = verifier.Verify(otherPayload, signature, signingCert)
	require.Error(t, err)
	assert.Equal(t, ErrInvalidSignature, err)

	err = verifier.Verify(payload, otherSignature, otherCert)
	require.Error(t, err)
	assert.Equal(t, ErrInvalidSignature, err)
}
