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

// TODO: this should be configurable through a property in the provider in the harvester conf file, based on the bundle TTL to be signed.
const signingCertTTL = 24 * 30 * time.Hour

var ErrInvalidSignature = errors.New("invalid signature")

// DiskSigner implements the Signer interface using a CA private key and CA cert stored on disk.
// It uses one-time-use keys to sign payloads, and returns the signed payload along with a new certificate signed by the caCert.
// Each signing operation generates a new private key and a new certificate signed by the CA private key.
type DiskSigner struct {
	caPrivateKey crypto.Signer
	caCert       *x509.Certificate

	// used for testing
	clk clock.Clock
}

// DiskVerifier implements the Verifier interface using a trust bundle stored on disk.
type DiskVerifier struct {
	trustBundle []*x509.Certificate

	// used for testing
	clk clock.Clock
}

// DiskSignerConfig is a configuration struct for creating a new DiskSigner
type DiskSignerConfig struct {
	// the path to the CA private key file
	CAPrivateKeyPath string
	// the path to the CA certificate file
	CACertPath string
	Clk        clock.Clock
}

// DiskVerifierConfig is a configuration struct for creating a new DiskVerifier
type DiskVerifierConfig struct {
	// the path to the public key file
	TrustBundlePath string
	Clk             clock.Clock
}

// Sign computes a signature for the given payload by first hashing it using SHA256, and returns
// the signature as a byte slice of bytes along with the certificate chain that has as a leaf the certificate
// of the public key that can be used to verify the signature.
func (s *DiskSigner) Sign(payload []byte) ([]byte, []*x509.Certificate, error) {
	now := s.clk.Now()

	// generate a new private key for signing
	key, err := cryptoutil.GenerateSigner(cryptoutil.DefaultKeyType)
	if err != nil {
		return nil, nil, err
	}

	serial, err := cryptoutil.NewSerialNumber()
	if err != nil {
		return nil, nil, err
	}
	// generate a new certificate for the public key signed by the CA private key
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: constants.Galadriel},
		NotBefore:             now,
		NotAfter:              now.Add(signingCertTTL),
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

	chain := []*x509.Certificate{cert, s.caCert}

	return signedPayload, chain, nil
}

// Verify checks if the signature of the given payload matches the expected signature.
// It also verifies that the certificate provided in the signature is signed by a trusted root CA.
func (v *DiskVerifier) Verify(payload, signature []byte, chain []*x509.Certificate) error {
	hashed := cryptoutil.CalculateDigest(payload)

	if len(chain) == 0 || chain[0] == nil {
		return fmt.Errorf("signing certificate is missing")
	}

	roots := x509.NewCertPool()
	for _, rootCert := range v.trustBundle {
		roots.AddCert(rootCert)
	}

	intermediates := x509.NewCertPool()
	for _, cert := range chain[1:] {
		intermediates.AddCert(cert)
	}

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   v.clk.Now(),
	}

	// verifies the root trust and intermediate certificates
	if _, err := chain[0].Verify(opts); err != nil {
		return fmt.Errorf("failed to verify chain: %v", err)
	}
	// verifies signature or artifact
	if err := rsa.VerifyPKCS1v15(chain[0].PublicKey.(*rsa.PublicKey), crypto.SHA256, hashed[:], signature); err != nil {
		return ErrInvalidSignature
	}

	return nil
}

// NewDiskSigner creates a new DiskSigner with the given configuration.
func NewDiskSigner(config *DiskSignerConfig) (*DiskSigner, error) {
	key, err := cryptoutil.LoadPrivateKey(config.CAPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load root CA private key: %w", err)
	}
	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, errors.New("root CA private key is not a signer")
	}

	cert, err := cryptoutil.LoadCertificate(config.CACertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load root CA certificate: %w", err)
	}

	if config.Clk == nil {
		config.Clk = clock.New()
	}

	return &DiskSigner{
		caPrivateKey: signer,
		caCert:       cert,
		clk:          config.Clk,
	}, nil
}

// NewDiskVerifier creates a new DiskVerifier with the given configuration
func NewDiskVerifier(config *DiskVerifierConfig) (*DiskVerifier, error) {
	certs, err := cryptoutil.LoadCertificates(config.TrustBundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load trust bundle: %w", err)
	}
	if config.Clk == nil {
		config.Clk = clock.New()
	}

	return &DiskVerifier{
		trustBundle: certs,
		clk:         config.Clk,
	}, nil
}

// NewDiskSignerConfig creates a new DiskSignerConfig function with the given CA private key and CA certificate paths
func NewDiskSignerConfig(CAPrivateKeyPath, CACertPath string) *DiskSignerConfig {
	return &DiskSignerConfig{
		CAPrivateKeyPath: CAPrivateKeyPath,
		CACertPath:       CACertPath,
		Clk:              clock.New(),
	}
}

// NewDiskVerifierConfig function creates a new DiskVerifierConfig function with the given trust bundle path
func NewDiskVerifierConfig(TrustBundlePath string) *DiskVerifierConfig {
	return &DiskVerifierConfig{
		TrustBundlePath: TrustBundlePath,
		Clk:             clock.New(),
	}
}
