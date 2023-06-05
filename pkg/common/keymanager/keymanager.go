package keymanager

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
)

// KeyManager provides a common interface for managing keys.
type KeyManager interface {
	// GenerateKey generates a new key with the given ID and key type.
	// If a key with that ID already exists, it is overwritten.
	GenerateKey(ctx context.Context, id string, keyType cryptoutil.KeyType) (Key, error)

	// GetKey returns the key with the given ID. If the key id does not exist,
	// an error is returned.
	GetKey(ctx context.Context, id string) (Key, error)

	// GetKeys returns all keys managed by the Memory.
	GetKeys(ctx context.Context) ([]Key, error)
}

// Key is an interface for an opaque key that can be used for signing.
// It also provides a method for getting the ID of the key.
type Key interface {
	ID() string
	Signer() crypto.Signer
}

// Generator is an interface for generating keys.
type Generator interface {
	GenerateRSA2048Key() (*rsa.PrivateKey, error)
	GenerateRSA4096Key() (*rsa.PrivateKey, error)
	// add method for more key types here
}

// defaultGenerator is a default implementation of Generator that uses
// crypto/rand.Reader as the source of entropy.
type defaultGenerator struct{}

// GenerateRSA2048Key generates a new RSA-2048 key.
func (b *defaultGenerator) GenerateRSA2048Key() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// GenerateRSA4096Key generates a new RSA-4096 key.
func (b *defaultGenerator) GenerateRSA4096Key() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 4096)
}
