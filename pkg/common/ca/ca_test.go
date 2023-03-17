package ca_test

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/ca"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/test/certtest"
	"github.com/golang-jwt/jwt/v4"
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
	config := newCAConfig(t, clock.NewFake())

	CA, err := ca.New(config)
	require.NoError(t, err)
	require.NotNil(t, CA)
}

func TestSignX509Certificate(t *testing.T) {
	config := newCAConfig(t, clock.NewFake())
	serverCA, err := ca.New(config)
	require.NoError(t, err)

	key, err := cryptoutil.CreateRSAKey()
	require.NoError(t, err)
	publicKey := key.Public()

	oneMinute := 1 * time.Minute

	params := ca.X509CertificateParams{
		PublicKey: publicKey,
		TTL:       oneMinute,
		Subject: pkix.Name{
			Organization: []string{"test-org"},
			CommonName:   "test-name",
		},
	}

	cert, err := serverCA.SignX509Certificate(params)
	require.NoError(t, err)
	require.NotNil(t, cert)

	// check the cert was signed by the serverCA
	err = cert.CheckSignatureFrom(config.RootCert)
	require.NoError(t, err)

	assert.NotNil(t, cert.SerialNumber)
	assert.Equal(t, []string{"test-org"}, cert.Subject.Organization)
	assert.Equal(t, "test-name", cert.Subject.CommonName)
	assert.Contains(t, cert.DNSNames, "test-name")
	assert.Equal(t, publicKey, cert.PublicKey)
	assert.False(t, cert.IsCA)
	assert.True(t, cert.BasicConstraintsValid)
	assert.Equal(t, config.Clock.Now().Add(ca.NotBeforeTolerance), cert.NotBefore)
	assert.Equal(t, config.Clock.Now().Add(oneMinute), cert.NotAfter)
	assert.Equal(t, cert.KeyUsage, expectedKeyUsage)
	assert.Equal(t, cert.ExtKeyUsage, expectedExtendedKeyUsage)
}

func TestSignJWT(t *testing.T) {
	config := newCAConfig(t, clock.New())
	serverCA, err := ca.New(config)
	require.NoError(t, err)

	oneMinute := 1 * time.Minute

	params := ca.JWTParams{
		Issuer:   "test-issuer",
		Subject:  spiffeid.RequireTrustDomainFromString("test-domain"),
		Audience: []string{"test-audience-1", "test-audience-2"},
		TTL:      oneMinute,
	}

	token, err := serverCA.SignJWT(params)
	require.NoError(t, err)
	require.NotNil(t, token)

	claims := &jwt.RegisteredClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) { return serverCA.PublicKey(), nil })

	require.NoError(t, err)
	require.NotNil(t, parsed)

	require.NoError(t, err)
	assert.Equal(t, claims.Issuer, "test-issuer")
	assert.Equal(t, claims.Subject, "test-domain")
	assert.Contains(t, claims.Audience, "test-audience-1")
	assert.Contains(t, claims.Audience, "test-audience-2")
	assert.Equal(t, claims.IssuedAt.Time.Unix(), config.Clock.Now().Unix())
	assert.Equal(t, claims.ExpiresAt.Time.Unix(), config.Clock.Now().Add(oneMinute).Unix())
}

func TestPublicKey(t *testing.T) {
	config := newCAConfig(t, clock.NewFake())
	serverCA, err := ca.New(config)
	require.NoError(t, err)

	assert.Equal(t, config.RootCert.PublicKey, serverCA.PublicKey())
}

func newCAConfig(t *testing.T, clk clock.Clock) *ca.Config {
	caCert, caKey, err := certtest.CreateTestCACertificate(clk)
	require.NoError(t, err)

	// success
	config := &ca.Config{
		RootCert: caCert,
		RootKey:  caKey,
		Clock:    clk,
	}

	return config
}
