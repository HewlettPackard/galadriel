package fakejwtissuer

import (
	"context"
	"crypto"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/jwt"
	"github.com/HewlettPackard/galadriel/test/jwttest"
	"github.com/jmhodges/clock"
	"github.com/stretchr/testify/require"
)

// JWTIssuer is a fake implementation of the JWTIssuer interface.
type JWTIssuer struct {
	Token  string
	Signer crypto.Signer
}

func New(t *testing.T, kid string, sub string, aud []string) *JWTIssuer {
	key, err := cryptoutil.GenerateSigner(cryptoutil.RSA2048)
	require.NoError(t, err)
	token := jwttest.CreateToken(t, clock.New(), key, kid, sub, "test", aud, time.Hour)
	return &JWTIssuer{
		Token:  token,
		Signer: key,
	}
}

func (j JWTIssuer) IssueJWT(ctx context.Context, params *jwt.JWTParams) (string, error) {
	return j.Token, nil
}
