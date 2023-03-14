package ca

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/test/certtest"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/jmhodges/clock"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	expectedKeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageKeyAgreement | x509.KeyUsageDigitalSignature
)

var (
	expectedExtendedKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
)

func TestNewCA(t *testing.T) {
	clk := clock.NewFake()
	caCert, caKey, err := certtest.CreateTestCACertificate(clk)
	require.NoError(t, err)

	// success
	config := &Config{
		RootCert: caCert,
		RootKey:  caKey,
		Clock:    clk,
	}

	ca, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, ca)
	assert.Equal(t, config.Clock, ca.clock)
}

func TestSignX509Certificate(t *testing.T) {
	clk := clock.NewFake()
	caCert, caKey, err := certtest.CreateTestCACertificate(clk)
	require.NoError(t, err)

	config := &Config{
		RootCert: caCert,
		RootKey:  caKey,
		Clock:    clk,
	}

	ca, _ := New(config)

	key, err := cryptoutil.CreateRSAKey()
	require.NoError(t, err)
	publicKey := key.Public()

	oneMinute := 1 * time.Minute

	params := X509CertificateParams{
		PublicKey: publicKey,
		TTL:       oneMinute,
		Subject: pkix.Name{
			Organization: []string{"test-org"},
			CommonName:   "test-name",
		},
	}

	cert, err := ca.SignX509Certificate(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, cert)

	assert.NotNil(t, cert.SerialNumber)
	assert.Equal(t, []string{"test-org"}, cert.Subject.Organization)
	assert.Equal(t, "test-name", cert.Subject.CommonName)
	assert.Contains(t, cert.DNSNames, "test-name")
	assert.Equal(t, publicKey, cert.PublicKey)
	assert.False(t, cert.IsCA)
	assert.True(t, cert.BasicConstraintsValid)
	assert.Equal(t, config.Clock.Now().Add(NotBeforeTolerance), cert.NotBefore)
	assert.Equal(t, config.Clock.Now().Add(oneMinute), cert.NotAfter)
	assert.Equal(t, cert.KeyUsage, expectedKeyUsage)
	assert.Equal(t, cert.ExtKeyUsage, expectedExtendedKeyUsage)
}

func TestSignJWT(t *testing.T) {
	clk := clock.NewFake()
	caCert, caKey, err := certtest.CreateTestCACertificate(clk)
	require.NoError(t, err)

	config := &Config{
		RootCert: caCert,
		RootKey:  caKey,
		Clock:    clk,
	}

	oneMinute := 1 * time.Minute

	ca, _ := New(config)

	params := JWTParams{
		Subject:  spiffeid.RequireTrustDomainFromString("domain.test"),
		Audience: []string{"aud1", "aud2"},
		TTL:      oneMinute,
	}

	token, err := ca.SignJWT(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, token)

	parsed, err := jwt.ParseSigned(token)
	require.NoError(t, err)
	require.NotNil(t, parsed)
	assert.Equal(t, ca.jwtCA.Kid, parsed.Headers[0].KeyID)

	publicKey := ca.PublicKey

	claims := make(map[string]any)
	err = parsed.Claims(publicKey, &claims)
	require.NoError(t, err)
	assert.Equal(t, claims["sub"], "domain.test")
	assert.Contains(t, claims["aud"], "aud1")
	assert.Contains(t, claims["aud"], "aud2")
	assert.Equal(t, claims["exp"], oneMinute.Seconds())
	assert.Equal(t, claims["iat"], float64(0))
}
