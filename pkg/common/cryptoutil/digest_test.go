package cryptoutil

import (
	"testing"
)

func TestValidateBundleDigest(t *testing.T) {
	payload := []byte("Hello World")
	digest := CalculateDigest(payload)

	err := ValidateBundleDigest(payload, digest)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateBundleDigest_InvalidDigest(t *testing.T) {
	payload := []byte("Hello World")
	invalidDigest := []byte("invalid_digest")

	err := ValidateBundleDigest(payload, invalidDigest)
	if err == nil {
		t.Error("expected an error, got nil")
	}
}
