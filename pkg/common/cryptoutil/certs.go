package cryptoutil

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
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
