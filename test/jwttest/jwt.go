package jwttest

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/jmhodges/clock"
	"github.com/stretchr/testify/require"
)

func CreateToken(t *testing.T, clk clock.Clock, privateKey crypto.Signer, kid, sub string, aud []string, ttl time.Duration) string {
	claims := map[string]interface{}{
		"sub": sub,
		"exp": jwt.NewNumericDate(clk.Now().Add(ttl)),
		"aud": aud,
		"iat": jwt.NewNumericDate(clk.Now()),
	}

	alg, err := cryptoutil.JoseAlgorithmFromPublicKey(privateKey.Public())
	require.NoError(t, err)

	jwtSigner, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: alg,
			Key: jose.JSONWebKey{
				Key:   privateKey,
				KeyID: kid,
			},
		},
		new(jose.SignerOptions).WithType("JWT"),
	)
	require.NoError(t, err)

	token, err := jwt.Signed(jwtSigner).Claims(claims).CompactSerialize()
	require.NoError(t, err)

	return token
}

func GenerateRandomKeyID(t *testing.T) string {
	keyIDBytes := make([]byte, 32)
	_, err := rand.Read(keyIDBytes)
	require.NoError(t, err)

	keyID := base64.RawURLEncoding.EncodeToString(keyIDBytes)
	return keyID
}
