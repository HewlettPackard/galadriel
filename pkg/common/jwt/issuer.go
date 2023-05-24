package jwt

import (
	"context"
	"crypto"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jmhodges/clock"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

const (
	kidHeader     = "kid"
	defaultJWTTTL = 10 * time.Minute
)

// Issuer is the interface used to sign JWTs.
type Issuer interface {
	// IssueJWT issues a JWT and returns the JWT.
	IssueJWT(context.Context, *JWTParams) (string, error)
}

// JWTParams holds the parameters for issuing a JWT.
type JWTParams struct {
	Issuer   string
	Subject  spiffeid.TrustDomain
	Audience []string
	TTL      time.Duration
}

// Config is the configuration for the JWTCA
type Config struct {
	// signer is an interface for an opaque private key
	// that can be used for signing operations
	Signer crypto.Signer

	// Kid is the id of the key used for signing
	Kid string
}

// JWTCA is an implementation of the Issuer interface that issues JWTs.
type JWTCA struct {
	// signer is an interface for an opaque private key
	// that can be used for signing operations
	signer crypto.Signer

	// kid is the id of the key used for signing
	kid string

	clk clock.Clock
}

// NewJWTCA creates a new JWTCA.
func NewJWTCA(c *Config) (*JWTCA, error) {
	if c.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if c.Kid == "" {
		return nil, fmt.Errorf("kid is required")
	}

	return &JWTCA{
		kid:    c.Kid,
		signer: c.Signer,
		clk:    clock.New(),
	}, nil
}

func (ca *JWTCA) IssueJWT(ctx context.Context, params *JWTParams) (string, error) {
	if params.TTL == 0 {
		params.TTL = defaultJWTTTL
	}
	expiresAt := ca.clk.Now().Add(params.TTL)
	now := ca.clk.Now()

	registeredClaims := jwt.RegisteredClaims{
		Issuer:    params.Issuer,
		Subject:   params.Subject.String(),
		Audience:  params.Audience,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, registeredClaims)
	token.Header[kidHeader] = ca.kid
	signedToken, err := token.SignedString(ca.signer)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}
