package disk

import (
	"context"
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/x509ca"
	"github.com/jmhodges/clock"
	"github.com/sirupsen/logrus"
)

// X509CA is a CA that signs X509 certificates using a disk-based private key and ROOT CA certificate.
type X509CA struct {
	// Interface for an opaque private key that can be used for signing operations.
	signer crypto.Signer

	// ROOT CA certificate for signing X509 certificates.
	certificate *x509.Certificate

	clock clock.Clock

	logger logrus.FieldLogger
}

// Config is the configuration for a disk-based X509CA.
type Config struct {
	// The path to the file containing the X.509 ROOT CA certificate.
	CertFilePath string `hcl:"cert_file_path"`
	// The path to the file containing the X.509 ROOT CA private key.
	KeyFilePath string `hcl:"key_file_path"`
}

// New creates a new disk-based X509CA.
// The returned X509CA is not configured.
// Call Configure() to configure it passing the HCL configuration.
func New() (*X509CA, error) {
	return &X509CA{
		clock:  clock.New(),
		logger: logrus.WithField(telemetry.SubsystemName, telemetry.DiskX509CA),
	}, nil
}

// Configure configures the disk-based X509CA from the given map.
func (ca *X509CA) Configure(config *Config) error {
	if config.CertFilePath == "" {
		return errors.New("certificate file path is required")
	}

	if config.KeyFilePath == "" {
		return errors.New("private key file path is required")
	}

	key, err := cryptoutil.LoadPrivateKey(config.KeyFilePath)
	if err != nil {
		return fmt.Errorf("unable to load private key: %v", err)
	}

	cert, err := cryptoutil.LoadCertificate(config.CertFilePath)
	if err != nil {
		return fmt.Errorf("unable to load certificate: %v", err)
	}

	// verify the certificate is self-signed (i.e. ROOT CA)
	if err := cert.CheckSignatureFrom(cert); err != nil {
		return fmt.Errorf("certificate is not self-signed")
	}

	// verify the certificate public key matches the private key
	if err := cryptoutil.VerifyCertificatePrivateKey(cert, key); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	ca.certificate = cert
	ca.signer = key.(crypto.Signer)

	return nil
}

// IssueX509Certificate issues an X509 certificate using the disk-based private key and ROOT CA certificate. The certificate
// is bound to the given public key and subject.
func (ca *X509CA) IssueX509Certificate(ctx context.Context, params *x509ca.X509CertificateParams) ([]*x509.Certificate, error) {
	if params.PublicKey == nil {
		return nil, errors.New("public key is required")
	}
	if params.TTL == 0 {
		return nil, errors.New("TTL is required")
	}

	template, err := cryptoutil.CreateX509Template(ca.clock, params.PublicKey, params.Subject, params.URIs, params.DNSNames, params.TTL)
	if err != nil {
		return nil, fmt.Errorf("failed to create template for Server certificate: %w", err)
	}

	cert, err := cryptoutil.SignX509(template, ca.certificate, ca.signer)
	if err != nil {
		return nil, fmt.Errorf("failed to sign X509 certificate: %w", err)
	}

	return []*x509.Certificate{cert}, nil
}
