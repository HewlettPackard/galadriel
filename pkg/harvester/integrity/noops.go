package integrity

import "crypto/x509"

// NoOpSigner is a no-op implementation of the Signer interface.
type NoOpSigner struct{}

func NewNoOpSigner() *NoOpSigner {
	return &NoOpSigner{}
}

// Sign computes a signature for the given payload and returns a no-op signature.
func (s NoOpSigner) Sign(payload []byte) ([]byte, []*x509.Certificate, error) {
	return nil, nil, nil
}

// NoOpVerifier is a no-op implementation of the Verifier interface.
type NoOpVerifier struct{}

func NewNoOpVerifier() *NoOpVerifier {
	return &NoOpVerifier{}
}

// Verify checks if the signature of the given payload matches the expected signature, which is always considered valid.
func (v NoOpVerifier) Verify(payload, signature []byte, certChain []*x509.Certificate) error {
	return nil
}
