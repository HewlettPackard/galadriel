package integrity

import "crypto/x509"

// Signer is an interface for signing payloads.
type Signer interface {
	// Sign computes a signature for the given payload and returns it as a byte slice, and optionally the certificate
	// used for signing along with intermediate certificates.
	Sign(payload []byte) ([]byte, []*x509.Certificate, error)
}

// Verifier is an interface for verifying signatures on payloads.
type Verifier interface {
	// Verify checks if the signature of the given payload matches the expected signature, using optionally a provided certificate chain for verification.
	Verify(payload, signature []byte, certChain []*x509.Certificate) error
}
