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
	"os"
	"time"

	"github.com/jmhodges/clock"
)

var (
	maxBigInt64 = getMaxBigInt64()
	one         = big.NewInt(1)
)

func CreateCertificate(template, parent *x509.Certificate, publicKey, signingKey any) (*x509.Certificate, error) {
	derBytes, err := x509.CreateCertificate(rand.Reader, template, parent, publicKey, signingKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(derBytes)
}

func LoadCertificate(path string) (*x509.Certificate, error) {
	certFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed reading certificate: %w", err)
	}

	return ParseCertificate(certFile)
}

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

func EncodeCertificate(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
}

func CreateCATemplate(clk clock.Clock, subject, issuer pkix.Name, ttl time.Duration) (*x509.Certificate, error) {
	now := clk.Now()
	serial, err := NewSerialNumber()
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber:          serial,
		Subject:               subject,
		Issuer:                issuer,
		IsCA:                  true,
		NotBefore:             now,
		NotAfter:              now.Add(ttl),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}, nil
}

func SelfSign(template *x509.Certificate) (*x509.Certificate, crypto.PrivateKey, error) {
	signerPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	certData, err := x509.CreateCertificate(rand.Reader, template, template, signerPrivateKey.Public(), signerPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, nil, err
	}

	return cert, signerPrivateKey, nil
}

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
