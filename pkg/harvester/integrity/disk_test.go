package integrity

import (
	"os"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/constants"
	"github.com/HewlettPackard/galadriel/test/certtest"
	"github.com/jmhodges/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var clk = clock.NewFake()

func TestNewDiskSigner(t *testing.T) {
	ds := NewDiskSigner()
	assert.NotNil(t, ds)
}

func TestNewDiskVerifier(t *testing.T) {
	dv := NewDiskVerifier()
	assert.NotNil(t, dv)
}

func TestSignAndVerifyUsingRootCA(t *testing.T) {
	tempDir := setupTest(t)

	diskSignerConfig := &DiskSignerConfig{
		CACertPath:       tempDir + "/root-ca.crt",
		CAPrivateKeyPath: tempDir + "/root-ca.key",
		TrustBundlePath:  "",
		SigningCertTTL:   "5m",
		Clock:            clk,
	}

	diskSigner := NewDiskSigner()
	err := diskSigner.Configure(diskSignerConfig)
	require.NoError(t, err)

	diskVerifierConfig := &DiskVerifierConfig{
		TrustBundlePath: tempDir + "/root-ca.crt",
		Clock:           clk,
	}
	diskVerifier := NewDiskVerifier()
	err = diskVerifier.Configure(diskVerifierConfig)
	require.NoError(t, err)

	// payload to sign
	payload := []byte("test payload")

	// Sign the payload using the disk signer
	signature, signingChain, err := diskSigner.Sign(payload)
	require.NoError(t, err)
	require.NotNil(t, signature)
	require.NotNil(t, signingChain)
	assert.Equal(t, 1, len(signingChain))

	// Verify the signature using the disk verifier
	err = diskVerifier.Verify(payload, signature, signingChain)
	require.NoError(t, err)

	alteredPayload := []byte("altered payload")
	err = diskVerifier.Verify(alteredPayload, signature, signingChain)
	require.Error(t, err)
	assert.Equal(t, ErrInvalidSignature, err)
}

func TestSignAndVerifyUsingIntermediateCAs(t *testing.T) {
	tempDir := setupTest(t)

	diskSignerConfig := &DiskSignerConfig{
		CACertPath:       tempDir + "/chain.crt",
		CAPrivateKeyPath: tempDir + "/intermediate-ca-2.key",
		TrustBundlePath:  tempDir + "/bundle.crt",
		SigningCertTTL:   "5m",
		Clock:            clk,
	}

	diskSigner := NewDiskSigner()
	err := diskSigner.Configure(diskSignerConfig)
	require.NoError(t, err)

	diskVerifierConfig := &DiskVerifierConfig{
		TrustBundlePath: tempDir + "/bundle.crt",
		Clock:           clk,
	}
	diskVerifier := NewDiskVerifier()
	err = diskVerifier.Configure(diskVerifierConfig)
	require.NoError(t, err)

	// payload to sign
	payload := []byte("test payload")

	// Sign the payload using the disk signer
	signature, signingChain, err := diskSigner.Sign(payload)
	require.NoError(t, err)
	require.NotNil(t, signature)
	require.NotNil(t, signingChain)
	assert.Equal(t, 3, len(signingChain))

	// Verify the signature using the disk verifier
	err = diskVerifier.Verify(payload, signature, signingChain)
	require.NoError(t, err)

	alteredPayload := []byte("altered payload")
	err = diskVerifier.Verify(alteredPayload, signature, signingChain)
	require.Error(t, err)
	assert.Equal(t, ErrInvalidSignature, err)

	// remove last cert from the chain so it doesn't chain back to the root
	signingChain = signingChain[:len(signingChain)-1]
	err = diskVerifier.Verify(payload, signature, signingChain)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to verify signing certificate chain")
}

func TestConfigureDiskSigner(t *testing.T) {
	tempDir := setupTest(t)

	testCases := []struct {
		name                 string
		config               DiskSignerConfig
		err                  string
		expectedBundleLength int
	}{
		{
			name: "WithRootCA",
			config: DiskSignerConfig{
				CAPrivateKeyPath: tempDir + "/root-ca.key",
				CACertPath:       tempDir + "/root-ca.crt",
				TrustBundlePath:  "",
				SigningCertTTL:   "42h",
			},
			expectedBundleLength: 0,
		},
		{
			name: "WithIntermediateCAAndRootCA",
			config: DiskSignerConfig{
				CAPrivateKeyPath: tempDir + "/intermediate-ca.key",
				CACertPath:       tempDir + "/intermediate-ca.crt",
				TrustBundlePath:  tempDir + "/root-ca.crt",
				SigningCertTTL:   "5h",
			},
			expectedBundleLength: 0,
		},
		{
			name: "WithIntermediateCAAndTrustBundle",
			config: DiskSignerConfig{
				CAPrivateKeyPath: tempDir + "/intermediate-ca-2.key",
				CACertPath:       tempDir + "/chain.crt",
				TrustBundlePath:  tempDir + "/bundle.crt",
				SigningCertTTL:   "12h",
			},
			expectedBundleLength: 1,
		},
		{
			name: "WithIntermediateCADontChainBack",
			config: DiskSignerConfig{
				CAPrivateKeyPath: tempDir + "/other-ca.key",
				CACertPath:       tempDir + "/other-ca.crt",
				TrustBundlePath:  tempDir + "/root-ca.crt",
				SigningCertTTL:   "42h",
			},
			err: "unable to chain the certificate to a trusted CA",
		},
		{
			name: "CertAndPrivateKeyNotMatch",
			config: DiskSignerConfig{
				CAPrivateKeyPath: tempDir + "/other-ca.key",
				CACertPath:       tempDir + "/root-ca.crt",
				SigningCertTTL:   "42h",
			},
			err: "certificate public key does not match private key",
		},
		{
			name: "NoSelfSigned",
			config: DiskSignerConfig{
				CAPrivateKeyPath: tempDir + "/intermediate-ca.key",
				CACertPath:       tempDir + "/intermediate-ca.crt",
				SigningCertTTL:   "42h",
			},
			err: constants.ErrTrustBundleRequired,
		},
		{
			name: "NoIntermediateCACert",
			config: DiskSignerConfig{
				CAPrivateKeyPath: tempDir + "/intermediate-ca.key",
				CACertPath:       "",
				SigningCertTTL:   "42h",
			},
			err: constants.ErrCertPathRequired,
		},
		{
			name: "NoIntermediateCAPrivateKey",
			config: DiskSignerConfig{
				CAPrivateKeyPath: "",
				CACertPath:       tempDir + "/intermediate-ca.crt",
				SigningCertTTL:   "42h",
			},
			err: constants.ErrPrivateKeyPathRequired,
		},
		{
			name: "EmptyTTL",
			config: DiskSignerConfig{
				CAPrivateKeyPath: tempDir + "/root-ca.key",
				CACertPath:       tempDir + "/root-ca.crt",
				SigningCertTTL:   "",
			},
			err: constants.ErrTTLRequired,
		},
		{
			name: "WithNonExistentCertificatePath",
			config: DiskSignerConfig{
				CAPrivateKeyPath: tempDir + "/root-ca.key",
				CACertPath:       tempDir + "/non-existent.crt",
				SigningCertTTL:   "1h",
			},
			err: "failed to load CA certificate",
		},
		{
			name: "WithNonExistentKeyPath",
			config: DiskSignerConfig{
				CAPrivateKeyPath: tempDir + "/non-existeng.key",
				CACertPath:       tempDir + "/root-ca.crt",
				SigningCertTTL:   "1h",
			},
			err: "failed to load CA private key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := NewDiskSigner()
			tc.config.Clock = clk

			err := ds.Configure(&tc.config)
			if tc.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedBundleLength, len(ds.upstreamChain))
				assert.Equal(t, parseDuration(t, tc.config.SigningCertTTL), ds.signingCertTTL)
			}
		})
	}
}

func TestConfigureDiskVerifier(t *testing.T) {
	tempDir := setupTest(t)

	testCases := []struct {
		name                 string
		config               DiskVerifierConfig
		err                  string
		expectedBundleLength int
	}{
		{
			name: "WithOneRootCA",
			config: DiskVerifierConfig{
				TrustBundlePath: tempDir + "/root-ca.crt",
			},
			expectedBundleLength: 1,
		},
		{
			name: "WithBundle",
			config: DiskVerifierConfig{
				TrustBundlePath: tempDir + "/bundle.crt",
			},
			expectedBundleLength: 2,
		},
		{
			name: "WithNonExistentPath",
			config: DiskVerifierConfig{
				TrustBundlePath: tempDir + "/non-existent.crt",
			},
			err: "failed to load trust bundle",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := NewDiskVerifier()
			tc.config.Clock = clk

			err := ds.Configure(&tc.config)
			if tc.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedBundleLength, len(ds.trustBundle))
			}
		})
	}
}

func parseDuration(t *testing.T, d string) time.Duration {
	duration, err := time.ParseDuration(d)
	require.NoError(t, err)
	return duration
}

func setupTest(t *testing.T) string {
	tempDir := certtest.CreateTestCACertificates(t, clk)
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	t.Cleanup(cleanup)

	return tempDir
}
