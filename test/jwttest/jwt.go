package jwttest

import (
	"context"
	"crypto"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/jwt"
	"github.com/HewlettPackard/galadriel/pkg/common/keymanager"
	"github.com/HewlettPackard/galadriel/test/keytest"
	gojwt "github.com/golang-jwt/jwt/v4"
	"github.com/jmhodges/clock"
	"github.com/stretchr/testify/require"
)

type FakeKeyManager struct {
	key crypto.Signer
}

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

func NewJWTIssuer() (jwt.Issuer, crypto.Signer) {
	signer := keytest.MustSignerRSA2048()
	ca, err := jwt.NewJWTCA(&jwt.Config{
		Signer: signer,
		Kid:    "test",
	})
	if err != nil {
		panic(err)
	}

	return ca, signer
}

func NewJWTValidator(signer crypto.Signer, expAud []string) jwt.Validator {
	km := FakeKeyManager{key: signer}
	return jwt.NewDefaultJWTValidator(&jwt.ValidatorConfig{
		KeyManager:       km,
		ExpectedAudience: expAud,
	})
}

func (f FakeKeyManager) GenerateKey(ctx context.Context, id string, keyType cryptoutil.KeyType) (keymanager.Key, error) {
	panic("not supposed to be called")
}

func (f FakeKeyManager) GetKey(ctx context.Context, id string) (keymanager.Key, error) {
	return &keymanager.KeyEntry{
		PrivateKey: f.key,
		PublicKey:  f.key.Public(),
	}, nil
}

func (f FakeKeyManager) GetKeys(ctx context.Context) ([]keymanager.Key, error) {
	return []keymanager.Key{&keymanager.KeyEntry{
		PrivateKey: f.key,
		PublicKey:  f.key.Public(),
	}}, nil
}
