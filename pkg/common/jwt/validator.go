package jwt

import (
	"context"
	"crypto"
	"errors"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/keymanager"
	"github.com/golang-jwt/jwt/v4"
)

// Validator validates JWT tokens using a public key.
type Validator interface {
	// ValidateToken ValidateJWT validates a JWT and returns the claims.
	ValidateToken(context.Context, string) (*jwt.RegisteredClaims, error)
}

type ValidatorConfig struct {
	// KeyManager is the key manager used to get the public key for validating the JWT.
	KeyManager       keymanager.KeyManager
	ExpectedAudience []string
}

type DefaultJWTValidator struct {
	keyManager       keymanager.KeyManager
	expectedAudience []string
}

func NewDefaultJWTValidator(c *ValidatorConfig) *DefaultJWTValidator {
	return &DefaultJWTValidator{
		keyManager:       c.KeyManager,
		expectedAudience: c.ExpectedAudience,
	}
}

func (v *DefaultJWTValidator) ValidateToken(ctx context.Context, token string) (*jwt.RegisteredClaims, error) {
	if token == "" {
		return nil, errors.New("token is empty")
	}

	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return v.getPublicKey(ctx, token)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse and validate token: %w", err)
	}

	// Validate the claims
	err = claims.Valid()
	if err != nil {
		return nil, err
	}

	if !v.validAudience(claims) {
		return nil, jwt.NewValidationError("token audience is invalid", jwt.ValidationErrorAudience)
	}
	return claims, nil
}

func (v *DefaultJWTValidator) validAudience(claims *jwt.RegisteredClaims) bool {
	for _, aud := range v.expectedAudience {
		ok := claims.VerifyAudience(aud, true)
		if !ok {
			return false
		}
	}
	return true
}

func (v *DefaultJWTValidator) getPublicKey(ctx context.Context, t *jwt.Token) (crypto.PublicKey, error) {
	kid, ok := t.Header[kidHeader].(string)
	if !ok {
		return nil, errors.New("missing kid header")
	}

	key, err := v.keyManager.GetKey(ctx, kid)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve public key for kid %q: %w", kid, err)
	}

	return key.Signer().Public(), nil
}
