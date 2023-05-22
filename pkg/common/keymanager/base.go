package keymanager

import (
	"context"
	"crypto"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
)

// base implementation of KeyManager that can be embedded into
// other KeyManager implementations (e.g. memory and disk).
type base struct {
	mu *sync.RWMutex

	generator Generator
	entries   map[string]*KeyEntry
}

// KeyEntry is a key entry in the KeyManager.
// It implements the Key interface.
type KeyEntry struct {
	PrivateKey crypto.Signer
	PublicKey  crypto.PublicKey
	id         string
}

// Config is the configuration for a base KeyManager.
type Config struct {
	// Optional Key Generator
	Generator Generator
}

// newBase creates a new base KeyManager.
func newBase(config *Config) *base {
	if config == nil {
		config = &Config{}
	}
	if config.Generator == nil {
		config.Generator = &defaultGenerator{}
	}
	return &base{
		mu:        &sync.RWMutex{},
		generator: config.Generator,
		entries:   make(map[string]*KeyEntry),
	}
}

// Public returns the public key corresponding to the private key of the KeyEntry.
func (k *KeyEntry) Public() crypto.PublicKey {
	return k.PublicKey
}

// Sign signs digest with the private key of the KeyEntry.
func (k *KeyEntry) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	return k.PrivateKey.Sign(rand, digest, opts)
}

// ID returns the ID of the KeyEntry.
func (k *KeyEntry) ID() string {
	return k.id
}

func (k *KeyEntry) Signer() crypto.Signer {
	return k.PrivateKey
}

// GenerateKey creates a new key pair and stores it in the KeyManager.
func (b *base) GenerateKey(ctx context.Context, keyID string, keyType cryptoutil.KeyType) (Key, error) {
	if keyID == "" {
		return nil, errors.New("key id is required")
	}
	if keyType == cryptoutil.KeyTypeUnset {
		return nil, errors.New("key type is required")
	}

	newEntry, err := b.generateKeyEntry(keyID, keyType)
	if err != nil {
		return nil, err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.entries[keyID] = newEntry

	return newEntry, nil
}

func (b *base) GetKey(ctx context.Context, id string) (Key, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	entry, ok := b.entries[id]
	if !ok {
		return nil, fmt.Errorf("no such key %q", id)
	}

	return entry, nil
}

func (b *base) GetKeys(ctx context.Context) ([]Key, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	keys := make([]Key, 0, len(b.entries))
	for _, entry := range b.entries {
		keys = append(keys, entry)
	}

	return keys, nil
}

func (b *base) generateKeyEntry(keyID string, keyType cryptoutil.KeyType) (*KeyEntry, error) {
	var err error
	var privateKey crypto.Signer

	switch keyType {
	case cryptoutil.RSA2048:
		privateKey, err = b.generator.GenerateRSA2048Key()
	case cryptoutil.RSA4096:
		privateKey, err = b.generator.GenerateRSA4096Key()
	default:
		return nil, fmt.Errorf("unable to generate key %q for unknown key type %q", keyID, keyType)
	}
	if err != nil {
		return nil, err
	}

	return &KeyEntry{
		PrivateKey: privateKey,
		PublicKey:  privateKey.Public(),
		id:         keyID,
	}, nil
}
