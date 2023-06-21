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

const (
	ErrCertPathRequired       = "certificate file path is required"
	ErrPrivateKeyPathRequired = "private key file path is required"
	ErrPublicKeyRequired      = "public key is required"
	ErrTTLRequired            = "TTL is required"
	ErrTrustBundleRequired    = "certificate is not self-signed. A trust bundle is required"
)

// X509CA is a CA that signs X509 certificates using a disk-based private key and ROOT CA certificate.
type X509CA struct {
	// signer is an interface for an opaque private key that can be used for signing operations.
	signer crypto.Signer

	// certificate is the CA certificate for signing X509 certificates.
	certificate *x509.Certificate

	// upstreamChain contains the intermediates certificates necessary to
	// chain back to the upstream trust bundle.
	upstreamChain []*x509.Certificate

	clock  clock.Clock
	logger logrus.FieldLogger
}

// Config is the configuration for a disk-based X509CA.
type Config struct {
	// The path to the file containing the X.509 CA certificate.
	CertFilePath string `hcl:"cert_file_path"`
	// The path to the file containing the X.509 CA private key.
	KeyFilePath string `hcl:"key_file_path"`
	// The path to the file containing the X.509 trust bundle.
	BundleFilePath string `hcl:"bundle_file_path"`

	Clock clock.Clock
}

// New creates a new disk-based X509CA.
// The returned X509CA is not configured.
// Call Configure() to configure it passing the HCL configuration.
func New() (*X509CA, error) {
	return &X509CA{
		logger: logrus.WithField(telemetry.SubsystemName, telemetry.DiskX509CA),
	}, nil
}

// Configure configures the disk-based X509CA from the given map.
func (ca *X509CA) Configure(config *Config) error {
	if config == nil {
		return errors.New("configuration is required")
	}

	if config.Clock == nil {
		config.Clock = clock.New()
	}
	ca.clock = config.Clock

	if config.CertFilePath == "" {
		return errors.New(ErrCertPathRequired)
	}

	if config.KeyFilePath == "" {
		return errors.New(ErrPrivateKeyPathRequired)
	}

	key, err := cryptoutil.LoadPrivateKey(config.KeyFilePath)
	if err != nil {
		return fmt.Errorf("unable to load private key: %v", err)
	}

	cert, err := cryptoutil.LoadCertificate(config.CertFilePath)
	if err != nil {
		return fmt.Errorf("unable to load certificate: %v", err)
	}

	if err := cryptoutil.VerifyCertificatePrivateKey(cert, key); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	if err := ca.processTrustBundle(config.BundleFilePath, cert); err != nil {
		return err
	}

	ca.certificate = cert
	ca.signer = key.(crypto.Signer)

	return nil
}

// IssueX509Certificate issues an X509 certificate using the disk-based private key and ROOT CA certificate. The certificate
// is bound to the given public key and subject.
func (ca *X509CA) IssueX509Certificate(ctx context.Context, params *x509ca.X509CertificateParams) ([]*x509.Certificate, error) {
	if params.PublicKey == nil {
		return nil, errors.New(ErrPublicKeyRequired)
	}
	if params.TTL == 0 {
		return nil, errors.New(ErrTTLRequired)
	}

	template, err := cryptoutil.CreateX509Template(ca.clock, params.PublicKey, params.Subject, params.URIs, params.DNSNames, params.TTL)
	if err != nil {
		return nil, fmt.Errorf("failed to create template for Server certificate: %w", err)
	}

	cert, err := cryptoutil.SignX509(template, ca.certificate, ca.signer)
	if err != nil {
		return nil, fmt.Errorf("failed to sign X509 certificate: %w", err)
	}

	chain, err := ca.buildCertificateChain(cert)
	if err != nil {
		return nil, fmt.Errorf("failed to build certificate chain: %w", err)
	}

	ca.logger.WithFields(logrus.Fields{
		"subject": cert.Subject,
		"expiry":  cert.NotAfter,
	}).Info("Successfully issued new X.509 certificate")

	return chain, nil
}

func (ca *X509CA) buildCertificateChain(leafCert *x509.Certificate) ([]*x509.Certificate, error) {
	chain := []*x509.Certificate{leafCert}

	// Always include the certificate used to sign the leaf certificate if it is an intermediate CA
	if !cryptoutil.IsSelfSigned(ca.certificate) {
		chain = append(chain, ca.certificate)
	}

	// If the CA has an upstream chain, append the intermediate certificates to the chain
	if len(ca.upstreamChain) > 0 {
		chain = append(chain, ca.upstreamChain...)
	}

	return chain, nil
}

func (ca *X509CA) processTrustBundle(bundlePath string, cert *x509.Certificate) error {
	if bundlePath == "" {
		return verifySelfSigned(cert)
	}

	bundle, err := cryptoutil.LoadCertificates(bundlePath)
	if err != nil {
		return fmt.Errorf("unable to load trust bundle: %v", err)
	}

	roots, intermediates := cryptoutil.SplitCertsIntoRootsAndIntermediates(bundle)
	if err := cryptoutil.VerifyCertificateChain([]*x509.Certificate{cert}, intermediates, roots, ca.clock.Now()); err != nil {
		return fmt.Errorf("certificate chain verification failed: %w", err)
	}

	// Remove potential duplicate: Ensure that the same certificate does not exist twice in the upstream chain.
	upstreamChain := cryptoutil.RemoveCertificateFromBundle(intermediates, cert)
	ca.upstreamChain = upstreamChain

	return nil
}

func verifySelfSigned(cert *x509.Certificate) error {
	if !cryptoutil.IsSelfSigned(cert) {
		return errors.New(ErrTrustBundleRequired)
	}
	return nil
}
