package cryptoutil

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"time"

	"github.com/jmhodges/clock"
)

const (
	// NotBeforeTolerance is used to allow for a small amount of clock skew when
	// validating the NotBefore field of a certificate.
	NotBeforeTolerance = 30 * time.Second

	certType = "CERTIFICATE"
)

var (
	maxBigInt64 = getMaxBigInt64()
	one         = big.NewInt(1)
)

// LoadCertificate loads a x509.Certificate from the given path.
func LoadCertificate(path string) (*x509.Certificate, error) {
	certFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed reading certificate: %w", err)
	}

	return ParseCertificate(certFile)
}

// LoadCertificates loads one or more certificates into an []*x509.Certificate from a PEM file.
func LoadCertificates(path string) ([]*x509.Certificate, error) {
	rest, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var certs []*x509.Certificate
	for blockno := 0; ; blockno++ {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != certType {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("unable to parse certificate in block %d: %w", blockno, err)
		}
		certs = append(certs, cert)
	}

	if len(certs) == 0 {
		return nil, errors.New("no certificates found in file")
	}

	return certs, nil
}

// ParseCertificate parses a x509.Certificate from the given PEM bytes.
func ParseCertificate(pemBytes []byte) (*x509.Certificate, error) {
	certPEM, _ := pem.Decode(pemBytes)
	if certPEM == nil {
		return nil, errors.New("failed decoding certificate PEM")
	}

	cert, err := x509.ParseCertificate(certPEM.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed parsing certificate: %w", err)
	}

	return cert, nil
}

// ParseCertificates parses a list of x509.Certificates from the given PEM bytes.
func ParseCertificates(pemBytes []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	block, rest := pem.Decode(pemBytes)
	for block != nil {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed parsing certificate: %w", err)
		}
		certs = append(certs, cert)
		block, rest = pem.Decode(rest)
	}
	return certs, nil
}

// EncodeCertificate encodes the given x509.Certificate into PEM format.
func EncodeCertificate(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: certType, Bytes: cert.Raw})
}

// CreateX509Template creates a new x509.Certificate template for a leaf certificate.
func CreateX509Template(clk clock.Clock, publicKey crypto.PublicKey, subject pkix.Name, uris []*url.URL, dnsNames []string, ttl time.Duration) (*x509.Certificate, error) {
	now := clk.Now()
	serial, err := NewSerialNumber()
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               subject,
		IsCA:                  false,
		NotBefore:             now.Add(-NotBeforeTolerance),
		NotAfter:              now.Add(ttl),
		BasicConstraintsValid: true,
		PublicKey:             publicKey,
		URIs:                  uris,
		DNSNames:              dnsNames,
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

// CreateCATemplate creates a new x509.Certificate template for a CA certificate.
func CreateCATemplate(clk clock.Clock, publicKey crypto.PublicKey, subject pkix.Name, ttl time.Duration) (*x509.Certificate, error) {
	now := clk.Now()
	serial, err := NewSerialNumber()
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber:          serial,
		Subject:               subject,
		IsCA:                  true,
		NotBefore:             now,
		NotAfter:              now.Add(ttl),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		PublicKey:             publicKey,
	}, nil
}

// CreateRootCATemplate creates a new x509.Certificate template for a root CA certificate.
func CreateRootCATemplate(clk clock.Clock, subject pkix.Name, ttl time.Duration) (*x509.Certificate, error) {
	now := clk.Now()
	serial, err := NewSerialNumber()
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber:          serial,
		Subject:               subject,
		IsCA:                  true,
		NotBefore:             now,
		NotAfter:              now.Add(ttl),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}, nil
}

// SignX509 creates a new x509.Certificate based on the given template.
// The parent certificate is the issuer of the new certificate.
// The signerPrivateKey is used to sign the new certificate.
func SignX509(template, parent *x509.Certificate, signerPrivateKey crypto.PrivateKey) (*x509.Certificate, error) {
	certData, err := x509.CreateCertificate(rand.Reader, template, parent, template.PublicKey, signerPrivateKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// SelfSignX509 creates a new self-signed x509.Certificate based on the given template.
// Returns the signed certificate and the private key used to sign it.
func SelfSignX509(template *x509.Certificate) (*x509.Certificate, crypto.PrivateKey, error) {
	signerPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	template.PublicKey = signerPrivateKey.Public()
	cert, err := SignX509(template, template, signerPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	return cert, signerPrivateKey, nil
}

// NewSerialNumber returns a new random serial number in the range [1, 2^63-1].
func NewSerialNumber() (*big.Int, error) {
	s, err := rand.Int(rand.Reader, maxBigInt64)
	if err != nil {
		return nil, fmt.Errorf("failed to create random number: %w", err)
	}

	return s.Add(s, one), nil
}

func getMaxBigInt64() *big.Int {
	return new(big.Int).SetInt64(1<<63 - 1)
}
