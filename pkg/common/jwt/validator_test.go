package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/keymanager"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expAud = []string{"test-audience-1", "test-audience-2"}

func Setup(t *testing.T) (*JWTCA, *DefaultJWTValidator) {
	ctx := context.Background()
	km := keymanager.New(&keymanager.Config{})
	config := ValidatorConfig{
		KeyManager:       km,
		ExpectedAudience: expAud,
	}
	jwtValidator := NewDefaultJWTValidator(&config)
	require.NotNil(t, jwtValidator)

	key, err := km.GenerateKey(ctx, "test-key-id", cryptoutil.RSA2048)
	require.NoError(t, err)
	require.NotNil(t, key)

	issuerConfig := &Config{
		Signer: key.Signer(),
		Kid:    "test-key-id",
	}
	jwtIssuer, err := NewJWTCA(issuerConfig)
	require.NoError(t, err)
	require.NotNil(t, jwtIssuer)

	return jwtIssuer, jwtValidator
}

func TestValidateTokenSuccess(t *testing.T) {
	ctx := context.Background()
	jwtIssuer, jwtValidator := Setup(t)

	params := &JWTParams{
		Issuer:   "test-issuer",
		Subject:  spiffeid.RequireTrustDomainFromString("test-subject"),
		Audience: expAud,
		TTL:      5 * time.Minute,
	}

	token, err := jwtIssuer.IssueJWT(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, token)

	// Test valid token
	claims, err := jwtValidator.ValidateToken(ctx, token)
	require.NoError(t, err)
	require.NotNil(t, claims)
}

func TestValidateInvalidAudience(t *testing.T) {
	ctx := context.Background()
	jwtIssuer, jwtValidator := Setup(t)

	params := &JWTParams{
		Issuer:   "test-issuer",
		Subject:  spiffeid.RequireTrustDomainFromString("test-subject"),
		Audience: []string{"other"},
		TTL:      5 * time.Minute,
	}

	token, err := jwtIssuer.IssueJWT(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, token)

	_, err = jwtValidator.ValidateToken(ctx, token)
	require.Error(t, err)
	assert.Equal(t, "token audience is invalid", err.Error())
}

func TestValidateKeyMisMatch(t *testing.T) {
	ctx := context.Background()
	jwtIssuer, jwtValidator := Setup(t)

	params := &JWTParams{
		Issuer:   "test-issuer",
		Subject:  spiffeid.RequireTrustDomainFromString("test-subject"),
		Audience: []string{"other"},
		TTL:      5 * time.Minute,
	}

	otherSigner, err := cryptoutil.GenerateSigner(cryptoutil.RSA2048)
	require.NoError(t, err)

	jwtIssuer.signer = otherSigner

	token, err := jwtIssuer.IssueJWT(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, token)

	_, err = jwtValidator.ValidateToken(ctx, token)
	require.Error(t, err)
	assert.Equal(t, "failed to parse and validate token: crypto/rsa: verification error", err.Error())
}
