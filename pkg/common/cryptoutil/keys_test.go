package cryptoutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/go-jose/go-jose/v3"
	"github.com/stretchr/testify/require"
)

func TestJoseAlgorithmFromPublicKey(t *testing.T) {
	var alg jose.SignatureAlgorithm
	var err error
	alg, err = JoseAlgorithmFromPublicKey(generateRSA(t, 1024).Public())
	require.NoError(t, err)
	require.Equal(t, alg, jose.RS256)

	alg, err = JoseAlgorithmFromPublicKey(generateRSA(t, 2048).Public())
	require.NoError(t, err)
	require.Equal(t, alg, jose.RS256)

	alg, err = JoseAlgorithmFromPublicKey(generateEC(t, elliptic.P521()).Public())
	require.EqualError(t, err, "unable to determine signature algorithm for public key type *ecdsa.PublicKey")
	require.Empty(t, alg)
}

func generateRSA(t *testing.T, bits int) *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	require.NoError(t, err)
	return privateKey
}

func generateEC(t *testing.T, curve elliptic.Curve) *ecdsa.PrivateKey {
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	require.NoError(t, err)
	return privateKey
}
