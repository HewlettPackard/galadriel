package disk

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/constants"
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
}

func TestConfigure(t *testing.T) {
	tempDir := setupTest(t)

	testCases := []struct {
		name                 string
		config               Config
		err                  string
		expectedBundleLength int
	}{
		{
			name: "WithRootCA",
			config: Config{
				KeyFilePath:  tempDir + "/root-ca.key",
				CertFilePath: tempDir + "/root-ca.crt",
			},
			expectedBundleLength: 0,
		},
		{
			name: "WithIntermediateCAAndRootCA",
			config: Config{
				KeyFilePath:    tempDir + "/intermediate-ca.key",
				CertFilePath:   tempDir + "/intermediate-ca.crt",
				BundleFilePath: tempDir + "/root-ca.crt",
			},
			expectedBundleLength: 0,
		},
		{
			name: "WithIntermediateChainAndTrustBundle",
			config: Config{
				KeyFilePath:    tempDir + "/intermediate-ca-2.key",
				CertFilePath:   tempDir + "/chain.crt",
				BundleFilePath: tempDir + "/bundle.crt",
			},
			expectedBundleLength: 1,
		},
		{
			name: "WithIntermediateCADontChainBackToRootCAInBundle",
			config: Config{
				KeyFilePath:    tempDir + "/intermediate-ca-2.key",
				CertFilePath:   tempDir + "/intermediate-ca-2.crt",
				BundleFilePath: tempDir + "/bundle.crt",
			},
			err: "unable to chain the certificate to a trusted CA",
		},
		{
			name: "CertAndPrivateKeyNotMatch",
			config: Config{
				KeyFilePath:  tempDir + "/other-ca.key",
				CertFilePath: tempDir + "/root-ca.crt",
			},
			err: "certificate public key does not match private key",
		},
		{
			name: "NoSelfSigned",
			config: Config{
				KeyFilePath:  tempDir + "/intermediate-ca.key",
				CertFilePath: tempDir + "/intermediate-ca.crt",
			},
			err: constants.ErrTrustBundleRequired,
		},
		{
			name: "NoIntermediateCACert",
			config: Config{
				KeyFilePath:  tempDir + "/intermediate-ca.key",
				CertFilePath: "",
			},
			err: constants.ErrCertPathRequired,
		},
		{
			name: "NoIntermediateCAPrivateKey",
			config: Config{
				KeyFilePath:  "",
				CertFilePath: tempDir + "/intermediate-ca.crt",
			},
			err: constants.ErrPrivateKeyPathRequired,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ca := newCA(t)
			tc.config.Clock = clk

			err := ca.Configure(&tc.config)
			if tc.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedBundleLength, len(ca.upstreamChain))
			}
		})
	}
}

func TestIssueX509CertificateUsingOnlyRootCA(t *testing.T) {
	tempDir := setupTest(t)
	config := Config{
		KeyFilePath:  tempDir + "/root-ca.key",
		CertFilePath: tempDir + "/root-ca.crt",
		Clock:        clk,
	}

	// Expectation: the certificate chain should only contain the leaf certificate when a Root CA is used.
	runIssueX509CertificateTest(t, config, 1)
}

func TestIssueX509CertificateWithIntermediateAndRootCA(t *testing.T) {
	tempDir := setupTest(t)
	config := Config{
		KeyFilePath:    tempDir + "/intermediate-ca.key",
		CertFilePath:   tempDir + "/intermediate-ca.crt",
		BundleFilePath: tempDir + "/root-ca.crt",
		Clock:          clk,
	}

	// Expectation: the certificate chain should contain the leaf certificate and the intermediate CA, but not the root CA when an Intermediate CA and a Root CA are used.
	runIssueX509CertificateTest(t, config, 2)
}

func TestIssueX509CertificateWithTwoIntermediateCAs(t *testing.T) {
	tempDir := setupTest(t)
	config := Config{
		KeyFilePath:    tempDir + "/intermediate-ca-2.key",
		CertFilePath:   tempDir + "/chain.crt",
		BundleFilePath: tempDir + "/bundle.crt",
		Clock:          clk,
	}

	// Expectation: the certificate chain should contain the leaf certificate and the two intermediate CAs when two Intermediate CAs are used.
	runIssueX509CertificateTest(t, config, 3)
}

func runIssueX509CertificateTest(t *testing.T, config Config, expectedChainLength int) {
	ca := newCA(t)

	err := ca.Configure(&config)
	require.NoError(t, err)

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
	require.Equal(t, expectedChainLength, len(certChain))

	leaf := certChain[0]
	require.Equal(t, dnsNames, leaf.DNSNames)
	require.Equal(t, uris, leaf.URIs)
	require.Equal(t, clk.Now().Add(fiveHours), leaf.NotAfter)
	require.Equal(t, clk.Now().Add(-cryptoutil.NotBeforeTolerance), leaf.NotBefore)
	require.Equal(t, expectedKeyUsage, leaf.KeyUsage)
	require.Equal(t, expectedExtKeyUsage, leaf.ExtKeyUsage)
	require.NotNil(t, leaf.SerialNumber)

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
	return ca
}

func setupTest(t *testing.T) string {
	tempDir := certtest.CreateTestCACertificates(t, clk)
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	t.Cleanup(cleanup)

	return tempDir
}
