package disk

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/x509ca"
	"github.com/HewlettPackard/galadriel/test/certtest"
	"github.com/jmhodges/clock"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	clk                 = clock.NewFake()
	uris                = []*url.URL{spiffeid.RequireFromString("spiffe://domain/test").URL()}
	dnsNames            = []string{"dns-test"}
	expectedKeyUsage    = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageKeyAgreement
	expectedExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	fiveHours           = time.Hour * 5
)

func TestNew(t *testing.T) {
	ca, err := New()
	require.NoError(t, err)
	assert.NotNil(t, ca)
	assert.NotNil(t, ca.clock)
}

func TestConfigure(t *testing.T) {
	tempDir, cleanup := setupTest(t)
	defer cleanup()

	config := Config{
		KeyFilePath:  tempDir + "/root-ca.key",
		CertFilePath: tempDir + "/root-ca.crt",
	}

	ca := newCA(t)
	err := ca.Configure(&config)
	require.NoError(t, err)
}

func TestConfigureCertAndPrivateKeyNotMatch(t *testing.T) {
	tempDir, cleanup := setupTest(t)
	defer cleanup()

	config := Config{
		KeyFilePath:  tempDir + "/other-ca.key",
		CertFilePath: tempDir + "/root-ca.crt",
	}

	ca := newCA(t)
	err := ca.Configure(&config)
	require.Error(t, err)
	assert.Equal(t, "certificate verification failed: certificate public key does not match private key", err.Error())
}

func TestConfigureNoSelfSigned(t *testing.T) {
	tempDir, cleanup := setupTest(t)
	defer cleanup()

	config := Config{
		KeyFilePath:  tempDir + "/intermediate-ca.key",
		CertFilePath: tempDir + "/intermediate-ca.crt",
	}

	ca := newCA(t)
	err := ca.Configure(&config)
	require.Error(t, err)
	assert.Equal(t, "certificate is not self-signed", err.Error())
}

func TestConfigureNoIntermediateCACert(t *testing.T) {
	tempDir, cleanup := setupTest(t)
	defer cleanup()

	config := Config{
		KeyFilePath:  tempDir + "/intermediate-ca.key",
		CertFilePath: "",
	}

	ca := newCA(t)

	err := ca.Configure(&config)
	require.Error(t, err)
	assert.Equal(t, "certificate file path is required", err.Error())
}

func TestConfigureNoIntermediateCAPrivateKey(t *testing.T) {
	tempDir, cleanup := setupTest(t)
	defer cleanup()

	config := Config{
		KeyFilePath:  "",
		CertFilePath: tempDir + "/intermediate-ca.crt",
	}

	ca := newCA(t)

	err := ca.Configure(&config)
	require.Error(t, err)
	assert.Equal(t, "private key file path is required", err.Error())
}

func TestIssueX509Certificate(t *testing.T) {
	tempDir, cleanup := setupTest(t)
	defer cleanup()

	config := Config{
		KeyFilePath:  tempDir + "/root-ca.key",
		CertFilePath: tempDir + "/root-ca.crt",
	}
	ca := newCA(t)

	err := ca.Configure(&config)
	require.NoError(t, err)

	// generate private key
	signer, err := cryptoutil.GenerateSigner(cryptoutil.RSA2048)
	require.NoError(t, err)

	params := &x509ca.X509CertificateParams{
		PublicKey: signer.Public(),
		Subject:   pkix.Name{CommonName: "test"},
		DNSNames:  dnsNames,
		URIs:      uris,
		TTL:       fiveHours,
	}
	certChain, err := ca.IssueX509Certificate(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, certChain)
	require.Len(t, certChain, 1)

	leaf := certChain[0]
	assert.Equal(t, dnsNames, leaf.DNSNames)
	assert.Equal(t, uris, leaf.URIs)
	assert.Equal(t, clk.Now().Add(fiveHours), leaf.NotAfter)
	assert.Equal(t, clk.Now().Add(-cryptoutil.NotBeforeTolerance), leaf.NotBefore)
	assert.Equal(t, expectedKeyUsage, leaf.KeyUsage)
	assert.Equal(t, expectedExtKeyUsage, leaf.ExtKeyUsage)
	assert.NotNil(t, leaf.SerialNumber)

	// verify that leaf certificate is signed by ca.Certificate
	x509CertPool := x509.NewCertPool()
	x509CertPool.AddCert(ca.certificate)
	opts := x509.VerifyOptions{
		Roots:       x509CertPool,
		CurrentTime: clk.Now(),
	}
	_, err = leaf.Verify(opts)
	require.NoError(t, err)
}

func newCA(t *testing.T) *X509CA {
	ca, err := New()
	require.NoError(t, err)
	assert.NotNil(t, ca)
	ca.clock = clock.NewFake()
	return ca
}

func setupTest(t *testing.T) (string, func()) {
	tempDir := certtest.CreateTestCACertificates(t, clk)
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}
