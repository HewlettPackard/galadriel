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
	clk := clock.NewFake()

	config := newCAConfig(t, clk)
	serverCA, err := ca.New(config)
	require.NoError(t, err)

	key, err := cryptoutil.CreateRSAKey()
	require.NoError(t, err)
	publicKey := key.Public()

	params := ca.X509CertificateParams{
		PublicKey: publicKey,
		TTL:       time.Minute,
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
	assert.Equal(t, params.Subject.Organization, cert.Subject.Organization)
	assert.Equal(t, params.Subject.CommonName, cert.Subject.CommonName)
	assert.Contains(t, cert.DNSNames, params.Subject.CommonName)
	assert.Equal(t, publicKey, cert.PublicKey)
	assert.False(t, cert.IsCA)
	assert.True(t, cert.BasicConstraintsValid)
	assert.Equal(t, clk.Now().Add(ca.NotBeforeTolerance), cert.NotBefore)
	assert.Equal(t, clk.Now().Add(params.TTL), cert.NotAfter)
	assert.Equal(t, cert.KeyUsage, expectedKeyUsage)
	assert.Equal(t, cert.ExtKeyUsage, expectedExtendedKeyUsage)
}

func TestSignJWT(t *testing.T) {
	clk := clock.NewFake()
	clk.Set(time.Now())

	config := newCAConfig(t, clk)
	serverCA, err := ca.New(config)
	require.NoError(t, err)

	params := ca.JWTParams{
		Issuer:   "test-issuer",
		Subject:  spiffeid.RequireTrustDomainFromString("test-domain"),
		Audience: []string{"test-audience-1", "test-audience-2"},
		TTL:      time.Minute,
	}

	token, err := serverCA.SignJWT(params)
	require.NoError(t, err)
	require.NotNil(t, token)

	claims := &jwt.RegisteredClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) { return serverCA.PublicKey(), nil })

	require.NoError(t, err)
	require.NotNil(t, parsed)

	c := &ca.Claims{
		RegisteredClaims: claims,
	}
	assertClaims(t, params, c, config.Clock)
}

func TestSignJWTWithCustomClaims(t *testing.T) {
	clk := clock.NewFake()
	clk.Set(time.Now())

	config := newCAConfig(t, clk)
	serverCA, err := ca.New(config)
	require.NoError(t, err)

	customClaims := &ca.CustomClaims{
		FederatesWith: "other.test",
	}

	params := ca.JWTParams{
		Issuer:       "test-issuer",
		Subject:      spiffeid.RequireTrustDomainFromString("test-domain"),
		Audience:     []string{"test-audience-1", "test-audience-2"},
		TTL:          1 * time.Minute,
		CustomClaims: customClaims,
	}

	token, err := serverCA.SignJWT(params)
	require.NoError(t, err)
	require.NotNil(t, token)

	claims := &ca.Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) { return serverCA.PublicKey(), nil })

	require.NoError(t, err)
	require.NotNil(t, parsed)

	assertClaims(t, params, claims, config.Clock)
	assert.Equal(t, params.CustomClaims.FederatesWith, claims.FederatesWith)
}

func assertClaims(t *testing.T, params ca.JWTParams, claims *ca.Claims, clk clock.Clock) {
	assert.Equal(t, params.Issuer, claims.Issuer)
	assert.Equal(t, params.Subject.String(), claims.Subject)
	assert.Equal(t, clk.Now().Unix(), claims.IssuedAt.Time.Unix())
	assert.Equal(t, clk.Now().Add(params.TTL).Unix(), claims.ExpiresAt.Time.Unix())
	assert.Equal(t, len(params.Audience), len(claims.Audience))
	for _, a := range params.Audience {
		assert.Contains(t, claims.Audience, a)
	}
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

	config := &ca.Config{
		RootCert: caCert,
		RootKey:  caKey,
		Clock:    clk,
	}

	return config
}
