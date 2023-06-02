package jwt

import (
	"context"
	"crypto"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/keymanager"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	expAud          = []string{"test-audience-1", "test-audience-2"}
	testTrustDomain = spiffeid.RequireTrustDomainFromString("test-subject")
)

func TestValidateToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		issuer     string
		audience   []string
		generate   func(*testing.T, *JWTCA, *DefaultJWTValidator, string, []string) string
		wantClaims bool
		wantError  string
	}{
		{
			name:     "valid token",
			issuer:   "test-issuer",
			audience: expAud,
			generate: func(t *testing.T, issuer *JWTCA, validator *DefaultJWTValidator, issuerName string, audience []string) string {
				params := &JWTParams{
					Issuer:   issuerName,
					Subject:  testTrustDomain,
					Audience: audience,
					TTL:      5 * time.Minute,
				}

				token, err := issuer.IssueJWT(ctx, params)
				require.NoError(t, err)
				require.NotNil(t, token)

				return token
			},
			wantClaims: true,
			wantError:  "",
		},
		{
			name:     "invalid audience",
			issuer:   "test-issuer",
			audience: []string{"other"},
			generate: func(t *testing.T, issuer *JWTCA, validator *DefaultJWTValidator, issuerName string, audience []string) string {
				params := &JWTParams{
					Issuer:   issuerName,
					Subject:  testTrustDomain,
					Audience: audience,
					TTL:      5 * time.Minute,
				}

				token, err := issuer.IssueJWT(ctx, params)
				require.NoError(t, err)
				require.NotNil(t, token)

				return token
			},
			wantClaims: false,
			wantError:  "token audience is invalid",
		},
		{
			name:     "mismatched public key",
			issuer:   "test-issuer",
			audience: expAud,
			generate: func(t *testing.T, issuer *JWTCA, validator *DefaultJWTValidator, issuerName string, audience []string) string {
				params := &JWTParams{
					Issuer:   issuerName,
					Subject:  testTrustDomain,
					Audience: audience,
					TTL:      5 * time.Minute,
				}

				otherSigner, err := cryptoutil.GenerateSigner(cryptoutil.RSA2048)
				require.NoError(t, err)
				issuer.signer = otherSigner

				token, err := issuer.IssueJWT(ctx, params)
				require.NoError(t, err)
				require.NotNil(t, token)

				return token
			},
			wantClaims: false,
			wantError:  "failed to parse and validate token: crypto/rsa: verification error",
		},
		{
			name:     "empty token",
			issuer:   "test-issuer",
			audience: expAud,
			generate: func(t *testing.T, issuer *JWTCA, validator *DefaultJWTValidator, issuerName string, audience []string) string {
				return ""
			},
			wantClaims: false,
			wantError:  "token is empty",
		},
		{
			name:     "missing kid header",
			issuer:   "test-issuer",
			audience: expAud,
			generate: func(t *testing.T, issuer *JWTCA, validator *DefaultJWTValidator, issuerName string, audience []string) string {
				signer, err := cryptoutil.GenerateSigner(cryptoutil.RSA2048)
				require.NoError(t, err)

				token := generateTestTokenWithoutKid(t, signer)
				require.NotNil(t, token)

				return token
			},
			wantClaims: false,
			wantError:  "failed to parse and validate token: missing kid header",
		},
		{
			name:     "unknown kid",
			issuer:   "test-issuer",
			audience: expAud,
			generate: func(t *testing.T, issuer *JWTCA, validator *DefaultJWTValidator, issuerName string, audience []string) string {
				signer, err := cryptoutil.GenerateSigner(cryptoutil.RSA2048)
				require.NoError(t, err)

				token := generateTestTokenUnknownKid(t, signer)
				require.NotNil(t, token)

				return token
			},
			wantClaims: false,
			wantError:  `failed to parse and validate token: failed to retrieve public key for kid "unknown-kid": no such key "unknown-kid"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issuer, validator := setup(t)

			token := tt.generate(t, issuer, validator, tt.issuer, tt.audience)
			claims, err := validator.ValidateToken(ctx, token)
			if tt.wantClaims {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			} else {
				assert.Error(t, err)
				assert.Nil(t, claims)
				assert.Contains(t, err.Error(), tt.wantError)
			}
		})
	}
}

func setup(t *testing.T) (*JWTCA, *DefaultJWTValidator) {
	ctx := context.Background()
	km := keymanager.NewMemoryKeyManager(nil)
	config := ValidatorConfig{
		KeyManager:       km,
		ExpectedAudience: expAud,
	}
	jwtValidator := NewDefaultJWTValidator(&config)
	require.NotNil(t, jwtValidator)

	key, err := km.GenerateKey(ctx, testKid, cryptoutil.RSA2048)
	require.NoError(t, err)
	require.NotNil(t, key)

	issuerConfig := &Config{
		Signer: key.Signer(),
		Kid:    testKid,
	}
	jwtIssuer, err := NewJWTCA(issuerConfig)
	require.NoError(t, err)
	require.NotNil(t, jwtIssuer)

	return jwtIssuer, jwtValidator
}

func generateTestTokenWithoutKid(t *testing.T, signer crypto.Signer) string {
	ctx := context.Background()
	jwtIssuer, _ := setup(t)

	params := &JWTParams{
		Issuer:   "test-issuer",
		Subject:  testTrustDomain,
		Audience: expAud,
		TTL:      5 * time.Minute,
	}

	token, err := jwtIssuer.IssueJWT(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, token)

	tokenObj, _ := jwt.Parse(token, nil)
	delete(tokenObj.Header, "kid")
	signedToken, _ := tokenObj.SignedString(signer)
	return signedToken
}

func generateTestTokenUnknownKid(t *testing.T, signer crypto.Signer) string {
	ctx := context.Background()
	jwtIssuer, _ := setup(t)

	params := &JWTParams{
		Issuer:   "test-issuer",
		Subject:  testTrustDomain,
		Audience: expAud,
		TTL:      5 * time.Minute,
	}

	token, err := jwtIssuer.IssueJWT(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, token)

	tokenObj, _ := jwt.Parse(token, nil)
	tokenObj.Header["kid"] = "unknown-kid"
	signedToken, _ := tokenObj.SignedString(signer)
	return signedToken
}
