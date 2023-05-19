package cryptoutil

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// CalculateDigest calculates the SHA256 digest of the given data.
func CalculateDigest(data []byte) []byte {
	digest := sha256.Sum256(data)
	return digest[:]
}

// ValidateBundleDigest validates the given payload against the given digest.
func ValidateBundleDigest(payload, digest []byte) error {
	calculatedDigest := CalculateDigest(payload)
	if !bytes.Equal(calculatedDigest, digest) {
		return fmt.Errorf("payload digest %q does not match calculated digest %q", digest, calculatedDigest)
	}

	return nil
}
