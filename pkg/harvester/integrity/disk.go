package integrity

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/constants"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/jmhodges/clock"
)

var ErrInvalidSignature = errors.New("invalid signature")

// DiskSigner implements the Signer interface using a CA private key and CA cert stored on disk.
// It uses one-time-use keys to sign payloads, and returns the signed payload along with a new certificate signed by the caCert.
// Each signing operation generates a new private key and a new certificate signed by the CA private key.
type DiskSigner struct {
	caPrivateKey crypto.Signer
	caCert       *x509.Certificate

	// upstreamChain contains the intermediates certificates necessary to
	// chain back to the upstream trust bundle.
	upstreamChain []*x509.Certificate

	signingCertTTL time.Duration

	// used for testing
	clock clock.Clock
}

// DiskVerifier implements the Verifier interface using a trust bundle stored on disk.
type DiskVerifier struct {
	trustBundle []*x509.Certificate

	// used for testing
	clock clock.Clock
}

// DiskSignerConfig is a configuration struct for creating a new DiskSigner
type DiskSignerConfig struct {
	CACertPath       string `hcl:"ca_cert_path"`
	CAPrivateKeyPath string `hcl:"ca_private_key_path"`
	TrustBundlePath  string `hcl:"trust_bundle_path"`
	SigningCertTTL   string `hcl:"signing_cert_ttl"`
	Clock            clock.Clock
}

// DiskVerifierConfig is a configuration struct for creating a new DiskVerifier
type DiskVerifierConfig struct {
	TrustBundlePath string `hcl:"trust_bundle_path"`
	Clock           clock.Clock
}

// NewDiskSigner creates a new DiskSigner instance without any configuration.
func NewDiskSigner() *DiskSigner {
	return &DiskSigner{}
}

// Configure sets up the DiskSigner with the provided configuration.
func (s *DiskSigner) Configure(config *DiskSignerConfig) error {
	if config == nil {
		return errors.New(constants.ErrConfigRequired)
	}

	var err error
	s.signingCertTTL, err = processTTL(config.SigningCertTTL)
	if err != nil {
		return err
	}

	if config.Clock == nil {
		config.Clock = clock.New()
	}
	s.clock = config.Clock

	if config.CACertPath == "" {
		return errors.New(constants.ErrCertPathRequired)
	}

	if config.CAPrivateKeyPath == "" {
		return errors.New(constants.ErrPrivateKeyPathRequired)
	}

	key, err := cryptoutil.LoadPrivateKey(config.CAPrivateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load CA private key: %w", err)
	}

	signer, ok := key.(crypto.Signer)
	if !ok {
		return errors.New("CA private key is not a signer")
	}

	s.caPrivateKey = signer

	certChain, err := cryptoutil.LoadCertificates(config.CACertPath)
	if err != nil {
		return fmt.Errorf("failed to load CA certificate: %w", err)
	}
	leafCert := certChain[0]

	if err := cryptoutil.VerifyCertificatePrivateKey(leafCert, key); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	s.caCert = leafCert

	if err := s.processTrustBundle(config.TrustBundlePath, certChain); err != nil {
		return err
	}

	return nil
}

// NewDiskVerifier creates a new DiskVerifier instance without any configuration.
func NewDiskVerifier() *DiskVerifier {
	return &DiskVerifier{}
}

// Configure sets up the DiskVerifier with the provided configuration.
func (v *DiskVerifier) Configure(config *DiskVerifierConfig) error {
	if config == nil {
		return errors.New(constants.ErrConfigRequired)
	}

	if config.Clock == nil {
		config.Clock = clock.New()
	}
	v.clock = config.Clock

	certs, err := cryptoutil.LoadCertificates(config.TrustBundlePath)
	if err != nil {
		return fmt.Errorf("failed to load trust bundle: %w", err)
	}

	if len(certs) == 0 {
		return errors.New("trust bundle must contain at least one certificate")
	}

	v.trustBundle = certs

	return nil
}

// Sign computes a signature for the given payload by first hashing it using SHA256, and returns
// the signature as a byte slice of bytes along with the certificate chain that has as a leaf the certificate
// of the public key that can be used to verify the signature.
func (s *DiskSigner) Sign(payload []byte) ([]byte, []*x509.Certificate, error) {
	now := s.clock.Now()

	// generate a new private key for signing
	key, err := cryptoutil.GenerateSigner(cryptoutil.DefaultKeyType)
	if err != nil {
		return nil, nil, err
	}

	serial, err := cryptoutil.NewSerialNumber()
	if err != nil {
		return nil, nil, err
	}

	// generate a certificate to bind the key used to sign the payload
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: constants.Galadriel},
		NotBefore:             now,
		NotAfter:              now.Add(s.signingCertTTL),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, s.caCert, key.Public(), s.caPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, err
	}

	hashedPayload := cryptoutil.CalculateDigest(payload)

	signedPayload, err := key.Sign(rand.Reader, hashedPayload[:], crypto.SHA256)
	if err != nil {
		return nil, nil, err
	}

	chain, err := s.buildCertificateChain(cert)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build signing certificate chain: %w", err)
	}

	return signedPayload, chain, nil
}

// Verify checks if the signature of the given payload matches the expected signature.
// It also verifies that the signing certificate chain can be chain back to a CA in the trust bundle.
func (v *DiskVerifier) Verify(payload, signature []byte, signingCertificateChain []*x509.Certificate) error {
	hashed := cryptoutil.CalculateDigest(payload)

	if len(signingCertificateChain) == 0 || signingCertificateChain[0] == nil {
		return errors.New("signing certificate chain is missing")
	}

	roots := v.trustBundle
	intermediates := signingCertificateChain[1:]
	err := cryptoutil.VerifyCertificateChain(signingCertificateChain, intermediates, roots, v.clock.Now())
	if err != nil {
		return fmt.Errorf("failed to verify signing certificate chain: %w", err)
	}

	// Verify the signature of the payload
	publicKey := signingCertificateChain[0].PublicKey.(*rsa.PublicKey)
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature); err != nil {
		return ErrInvalidSignature
	}

	return nil
}

func (s *DiskSigner) buildCertificateChain(cert *x509.Certificate) ([]*x509.Certificate, error) {
	chain := []*x509.Certificate{cert}

	// Always include the certificate used to sign the leaf certificate if it is an intermediate CA
	if !cryptoutil.IsSelfSigned(s.caCert) {
		chain = append(chain, s.caCert)
	}

	// If the CA has a upstreamChain, append the intermediate certificates in the trust bundle to the chain
	if len(s.upstreamChain) > 0 {
		chain = append(chain, s.upstreamChain...)
	}

	return chain, nil
}

func (s *DiskSigner) processTrustBundle(trustBundlePath string, certChain []*x509.Certificate) error {
	leafCert := certChain[0]
	if trustBundlePath == "" {
		return verifySelfSigned(leafCert)
	}

	bundle, err := cryptoutil.LoadCertificates(trustBundlePath)
	if err != nil {
		return fmt.Errorf("unable to load trust bundle: %w", err)
	}

	intermediates := certChain[1:]
	if err := cryptoutil.VerifyCertificateChain([]*x509.Certificate{leafCert}, intermediates, bundle, s.clock.Now()); err != nil {
		return fmt.Errorf("certificate chain verification failed: %w", err)
	}

	s.upstreamChain = intermediates

	return nil
}

func verifySelfSigned(cert *x509.Certificate) error {
	if !cryptoutil.IsSelfSigned(cert) {
		return errors.New(constants.ErrTrustBundleRequired)
	}
	return nil
}

func processTTL(ttl string) (time.Duration, error) {
	if ttl == "" {
		return 0, errors.New(constants.ErrTTLRequired)
	}
	duration, err := time.ParseDuration(ttl)
	if err != nil {
		return 0, fmt.Errorf("failed to parse signing cert TTL: %w", err)
	}
	if duration <= 0 {
		return 0, errors.New("signing cert TTL must be greater than 0")
	}

	return duration, nil
}
