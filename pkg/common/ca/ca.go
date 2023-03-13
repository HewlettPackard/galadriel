package ca

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/cryptosigner"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/jmhodges/clock"
)

var (
	maxBigInt64 = getMaxBigInt64()
	one         = big.NewInt(1)
)

// NotBeforeTolerance adds a margin to the NotBefore in case there is a clock drift across the servers
const NotBeforeTolerance = 30 * time.Second

type ServerCA interface {
	SignX509Certificate(ctx context.Context, params X509CertificateParams) (*x509.Certificate, error)
	SignJWT(ctx context.Context, params JWTParams) (string, error)
	GetPublicKey() crypto.PublicKey
}

type CA struct {
	PublicKey crypto.PublicKey

	x509CA *X509CA
	jwtCA  *JWTCA
	clock  clock.Clock
}

type Config struct {
	RootCertFile string
	RootKeyFile  string
	Clock        clock.Clock
}

func New(c *Config) (*CA, error) {
	cert, err := cryptoutil.LoadCertificate(c.RootCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed loading CA root certificate: %w", err)
	}

	key, err := cryptoutil.LoadRSAPrivateKey(c.RootKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed loading CA root private: %w", err)
	}

	signer, err := signerFromPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer from private key: %w", err)
	}

	x509CA := &X509CA{
		Certificate: cert,
		Signer:      signer,
	}

	kid, err := generateRandomKeyID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate random kid: %w", err)
	}

	jwtCA := &JWTCA{
		Signer: signer,
		Kid:    kid,
	}

	return &CA{
		x509CA:    x509CA,
		jwtCA:     jwtCA,
		clock:     c.Clock,
		PublicKey: signer.Public(),
	}, err
}

type X509CA struct {
	// CA certificate.
	Certificate *x509.Certificate

	// Signer is an interface for an opaque private key
	// that can be used for signing operations
	Signer crypto.Signer
}

type JWTCA struct {
	// Kid is the JWT 'kid' claim
	Kid string

	// Signer is an interface for an opaque private key
	// that can be used for signing operations
	Signer crypto.Signer
}

type X509CertificateParams struct {
	PublicKey crypto.PublicKey
	Subject   pkix.Name
	TTL       time.Duration
}

type JWTParams struct {
	Subject  string
	Audience []string
	TTL      time.Duration
}

func (ca *CA) GetPublicKey() crypto.PublicKey {
	return ca.PublicKey
}

func (ca *CA) SignX509Certificate(ctx context.Context, params X509CertificateParams) (*x509.Certificate, error) {
	template, err := ca.createX509Template(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create template for GCA certificate: %w", err)
	}

	cert, err := cryptoutil.CreateCertificate(template, ca.x509CA.Certificate, params.PublicKey, ca.x509CA.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to sign GCA certificate: %w", err)
	}

	return cert, nil
}

func (ca *CA) SignJWT(ctx context.Context, params JWTParams) (string, error) {
	expiresAt := ca.clock.Now().Add(params.TTL)
	now := ca.clock.Now()

	claims := map[string]interface{}{
		"sub": params.Subject,
		"exp": jwt.NewNumericDate(expiresAt),
		"aud": params.Audience,
		"iat": jwt.NewNumericDate(now),
	}

	alg, err := cryptoutil.JoseAlgorithmFromPublicKey(ca.jwtCA.Signer.Public())
	if err != nil {
		return "", fmt.Errorf("failed to determine JWT key algorithm: %w", err)
	}

	jwtSigner, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: alg,
			Key: jose.JSONWebKey{
				Key:   cryptosigner.Opaque(ca.jwtCA.Signer),
				KeyID: ca.jwtCA.Kid,
			},
		},
		new(jose.SignerOptions).WithType("JWT"),
	)
	if err != nil {
		return "", fmt.Errorf("failed to configure JWT signer: %w", err)
	}

	signedToken, err := jwt.Signed(jwtSigner).Claims(claims).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT SVID: %w", err)
	}

	return signedToken, nil
}

func (ca *CA) createX509Template(params X509CertificateParams) (*x509.Certificate, error) {
	serial, err := newSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to create X509 Template: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   params.Subject.CommonName,
			Organization: params.Subject.Organization,
		},
		NotBefore:             ca.clock.Now().Add(-NotBeforeTolerance),
		NotAfter:              ca.clock.Now().Add(params.TTL),
		BasicConstraintsValid: true,
		IsCA:                  false,
		PublicKey:             params.PublicKey,
		DNSNames:              []string{params.Subject.CommonName},
	}

	template.KeyUsage = x509.KeyUsageKeyEncipherment |
		x509.KeyUsageKeyAgreement |
		x509.KeyUsageDigitalSignature
	template.ExtKeyUsage = []x509.ExtKeyUsage{
		x509.ExtKeyUsageServerAuth,
		x509.ExtKeyUsageClientAuth,
	}
	return template, nil
}

func signerFromPrivateKey(privateKey crypto.PrivateKey) (crypto.Signer, error) {
	signer, ok := privateKey.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("expected crypto.Signer; got %T", privateKey)
	}
	return signer, nil
}

func newSerialNumber() (*big.Int, error) {
	s, err := rand.Int(rand.Reader, maxBigInt64)
	if err != nil {
		return nil, fmt.Errorf("failed to create random number: %w", err)
	}

	return s.Add(s, one), nil
}

func getMaxBigInt64() *big.Int {
	return new(big.Int).SetInt64(1<<63 - 1)
}

func generateRandomKeyID() (string, error) {
	// Generate 32 random bytes
	keyIDBytes := make([]byte, 32)
	_, err := rand.Read(keyIDBytes)
	if err != nil {
		return "", err
	}

	// Encode the bytes as a base64 string
	keyID := base64.RawURLEncoding.EncodeToString(keyIDBytes)
	return keyID, nil
}
