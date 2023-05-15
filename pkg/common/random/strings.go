package random

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateRandomString generates a random string of n bytes
func GenerateRandomString(nBytes uint) (string, error) {
	// Generate n random bytes
	randomStrBytes := make([]byte, nBytes)
	_, err := rand.Read(randomStrBytes)
	if err != nil {
		return "", err
	}

	randomStr := base64.RawURLEncoding.EncodeToString(randomStrBytes)
	return randomStr, nil
}
