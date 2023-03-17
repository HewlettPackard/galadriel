package ca

import (
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"sync"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmhodges/clock"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

// NotBeforeTolerance adds a margin to the NotBefore in case there is a clock drift across the servers
const NotBeforeTolerance = -30 * time.Second

type ServerCA interface {
	SignX509Certificate(params X509CertificateParams) (*x509.Certificate, error)
	SignJWT(params JWTParams) (string, error)
	PublicKey() crypto.PublicKey
}

type CA struct {
	mu sync.RWMutex

	publicKey crypto.PublicKey
	x509CA    *X509CA
	jwtCA     *JWTCA
	clock     clock.Clock

	Logger logrus.FieldLogger
}

type Config struct {
	RootCert *x509.Certificate
	RootKey  crypto.PrivateKey
	Clock    clock.Clock
	Logger   logrus.FieldLogger
}

type X509CA struct {
	CACertificate *x509.Certificate

	// Signer is an interface for an opaque private key
	// that can be used for signing operations
	Signer crypto.Signer
}

type JWTCA struct {
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
	Issuer   string
	Subject  spiffeid.TrustDomain
	Audience []string
	TTL      time.Duration
}

func New(c *Config) (*CA, error) {
	signer, err := signerFromPrivateKey(c.RootKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer from private key: %w", err)
	}

	x509CA := &X509CA{
		CACertificate: c.RootCert,
		Signer:        signer,
	}

	jwtCA := &JWTCA{
		Signer: signer,
	}

	return &CA{
		publicKey: signer.Public(),
		Logger:    c.Logger,
		x509CA:    x509CA,
		jwtCA:     jwtCA,
		clock:     c.Clock,
	}, err
}

func (ca *CA) SignX509Certificate(params X509CertificateParams) (*x509.Certificate, error) {
	x509CA := ca.getX509CA()

	template, err := ca.createX509Template(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create template for GCA certificate: %w", err)
	}

	cert, err := cryptoutil.CreateCertificate(template, x509CA.CACertificate, params.PublicKey, x509CA.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to sign GCA certificate: %w", err)
	}

	return cert, nil
}

func (ca *CA) SignJWT(params JWTParams) (string, error) {
	jwtCA := ca.getJWTCA()

	expiresAt := ca.clock.Now().Add(params.TTL)
	now := ca.clock.Now()

	claims := jwt.RegisteredClaims{
		Issuer:    params.Issuer,
		Subject:   params.Subject.String(),
		Audience:  params.Audience,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(jwtCA.Signer)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (ca *CA) PublicKey() crypto.PublicKey {
	return ca.publicKey
}

func (ca *CA) getX509CA() *X509CA {
	ca.mu.RLock()
	defer ca.mu.RUnlock()
	return ca.x509CA
}

func (ca *CA) getJWTCA() *JWTCA {
	ca.mu.RLock()
	defer ca.mu.RUnlock()
	return ca.jwtCA
}

func (ca *CA) createX509Template(params X509CertificateParams) (*x509.Certificate, error) {
	serial, err := cryptoutil.NewSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to create X509 Template: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   params.Subject.CommonName,
			Organization: params.Subject.Organization,
		},
		NotBefore:             ca.clock.Now().Add(NotBeforeTolerance),
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
