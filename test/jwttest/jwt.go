package jwttest

import (
	"crypto"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jmhodges/clock"
	"github.com/stretchr/testify/require"
)

func CreateToken(t *testing.T, clk clock.Clock, privateKey crypto.Signer, sub, iss string, aud []string, ttl time.Duration) string {
	claims := jwt.RegisteredClaims{
		Issuer:    iss,
		Subject:   sub,
		Audience:  aud,
		ExpiresAt: jwt.NewNumericDate(clk.Now().Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(clk.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	require.NoError(t, err)

	return signedToken
}
