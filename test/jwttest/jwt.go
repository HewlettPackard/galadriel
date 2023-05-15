package jwttest

import (
	"crypto"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/jwt"
	"github.com/HewlettPackard/galadriel/test/fakes/fakekeymanager"
	gojwt "github.com/golang-jwt/jwt/v4"
	"github.com/jmhodges/clock"
	"github.com/stretchr/testify/require"
)

func CreateToken(t *testing.T, clk clock.Clock, privateKey crypto.Signer, kid, sub, iss string, aud []string, ttl time.Duration) string {
	claims := gojwt.RegisteredClaims{
		Issuer:    iss,
		Subject:   sub,
		Audience:  aud,
		ExpiresAt: gojwt.NewNumericDate(clk.Now().Add(ttl)),
		IssuedAt:  gojwt.NewNumericDate(clk.Now()),
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signedToken, err := token.SignedString(privateKey)
	require.NoError(t, err)

	return signedToken
}

func NewJWTValidator(signer crypto.Signer, expAud []string) jwt.Validator {
	km := fakekeymanager.KeyManager{Key: signer}
	return jwt.NewDefaultJWTValidator(&jwt.ValidatorConfig{
		KeyManager:       km,
		ExpectedAudience: expAud,
	})
}
