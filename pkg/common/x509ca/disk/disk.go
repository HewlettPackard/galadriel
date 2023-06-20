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
	ErrCertPathRequired  = "certificate file path is required"
	ErrPrivateKeyPathReq = "private key file path is required"
	ErrPublicKeyRequired = "public key is required"
	ErrTTLRequired       = "TTL is required"
	ErrTrustBundleReq    = "certificate is not self-signed. A trust bundle is required"
)

// X509CA is a CA that signs X509 certificates using a disk-based private key and ROOT CA certificate.
type X509CA struct {
	// signer is an interface for an opaque private key that can be used for signing operations.
	signer crypto.Signer

	// certificate is the CA certificate for signing X509 certificates.
	certificate *x509.Certificate

	// trustBundle is a collection of trusted certificates.
	trustBundle []*x509.Certificate

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
		return errors.New(ErrPrivateKeyPathReq)
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

	if err := ca.processTrustBundle(config, cert); err != nil {
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

	return chain, nil
}

func (ca *X509CA) buildCertificateChain(cert *x509.Certificate) ([]*x509.Certificate, error) {
	chain := []*x509.Certificate{cert}

	// If the CA has a trust bundle, append the intermediate and the certificates in the trust bundle to the chain
	if len(ca.trustBundle) > 0 {
		chain = append(chain, ca.certificate)
		chain = append(chain, ca.trustBundle...)
	}

	return chain, nil
}

func (ca *X509CA) processTrustBundle(config *Config, cert *x509.Certificate) error {
	if config.BundleFilePath == "" {
		return ca.verifySelfSigned(cert)
	}

	return ca.loadAndVerifyTrustBundle(config, cert)
}

func (ca *X509CA) loadAndVerifyTrustBundle(config *Config, cert *x509.Certificate) error {
	bundle, err := ca.loadTrustBundle(config)
	if err != nil {
		return err
	}

	return ca.verifyTrustBundle(cert, bundle)
}

func (ca *X509CA) loadTrustBundle(config *Config) ([]*x509.Certificate, error) {
	bundle, err := cryptoutil.LoadCertificates(config.BundleFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to load trust bundle: %v", err)
	}

	return bundle, nil
}

func (ca *X509CA) verifyTrustBundle(cert *x509.Certificate, bundle []*x509.Certificate) error {
	if err := ca.verifyCertificateCanChainToTrustBundle(cert, bundle); err != nil {
		return fmt.Errorf("certificate chain verification failed: %w", err)
	}

	ca.trustBundle = bundle
	return nil
}

func (ca *X509CA) verifySelfSigned(cert *x509.Certificate) error {
	if err := cert.CheckSignatureFrom(cert); err != nil {
		return errors.New(ErrTrustBundleReq)
	}
	return nil
}

func (ca *X509CA) verifyCertificateCanChainToTrustBundle(cert *x509.Certificate, intermediates []*x509.Certificate) error {
	intermediatePool := x509.NewCertPool()
	rootPool := x509.NewCertPool()
	for _, intermediate := range intermediates {
		intermediatePool.AddCert(intermediate)
		rootPool.AddCert(intermediate)
	}

	opts := x509.VerifyOptions{
		Roots:         rootPool,
		Intermediates: intermediatePool,
		CurrentTime:   ca.clock.Now(),
	}

	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("unable to chain the certificate to a trusted CA: %w", err)
	}

	return nil
}
