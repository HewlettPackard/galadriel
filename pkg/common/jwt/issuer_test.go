package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKid    = "test-kid"
	testIssuer = "test-issuer"
)

func TestNew(t *testing.T) {
	signer, err := cryptoutil.GenerateSigner(cryptoutil.RSA2048)
	require.NoError(t, err)

	ca, err := NewJWTCA(&Config{
		Signer: signer,
		Kid:    testKid,
	})
	require.NoError(t, err)
	require.NotNil(t, ca)
	assert.Equal(t, testKid, ca.kid)
	assert.Equal(t, signer, ca.signer)
	assert.NotNil(t, ca.clk)

	// missing signer
	ca, err = NewJWTCA(&Config{
		Kid: testKid,
	})
	require.Error(t, err)
	require.Nil(t, ca)
	assert.Equal(t, "signer is required", err.Error())

	// missing kid
	ca, err = NewJWTCA(&Config{
		Signer: signer,
	})
	require.Error(t, err)
	require.Nil(t, ca)
	assert.Equal(t, "kid is required", err.Error())
}

func TestJWTCAIssueJWT(t *testing.T) {
	signer, err := cryptoutil.GenerateSigner(cryptoutil.RSA2048)
	require.NoError(t, err)

	ca, err := NewJWTCA(&Config{
		Signer: signer,
		Kid:    testKid,
	})
	require.NoError(t, err)

	params := &JWTParams{
		Issuer:   testIssuer,
		Subject:  spiffeid.RequireTrustDomainFromString("test-domain"),
		Audience: []string{"test-audience-1", "test-audience-2"},
		TTL:      time.Minute,
	}

	token, err := ca.IssueJWT(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, token)

	claims := &jwt.RegisteredClaims{}
	getServerPublicKey := func(token *jwt.Token) (interface{}, error) { return ca.signer.Public(), nil }
	parsed, err := jwt.ParseWithClaims(token, claims, getServerPublicKey)

	require.NoError(t, err)
	require.NotNil(t, parsed)

	// convert claims.Audience to a slice of strings
	audience := []string{}
	for _, a := range claims.Audience {
		audience = append(audience, a)
	}

	assert.Equal(t, ca.kid, parsed.Header["kid"])
	assert.Equal(t, params.Issuer, claims.Issuer)
	assert.Equal(t, params.Subject.String(), claims.Subject)
	assert.Equal(t, params.Audience, audience)
	assert.Equal(t, jwt.NewNumericDate(ca.clk.Now()), claims.IssuedAt)
}
