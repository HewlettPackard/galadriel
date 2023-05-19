package keymanager

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
)

// Disk extends the Base KeyManager to store keys in disk.
type Disk struct {
	Base

	keysFilePath string
}

// NewDiskKeyManager creates a new Disk that stores keys in disk.
func NewDiskKeyManager(generator Generator, keysFilePath string) (*Disk, error) {
	c := &Config{
		Generator: generator,
	}
	base := New(c)

	diskKeyManager := &Disk{
		Base:         *base,
		keysFilePath: keysFilePath,
	}

	err := diskKeyManager.loadKeysFromDisk()
	if err != nil {
		return nil, err
	}

	return diskKeyManager, nil
}

// GenerateKey generates a new key and stores it in disk.
func (d *Disk) GenerateKey(ctx context.Context, keyID string, keyType cryptoutil.KeyType) (Key, error) {
	key, err := d.Base.GenerateKey(ctx, keyID, keyType)
	if err != nil {
		return nil, err
	}

	err = d.saveKeysToDisk()
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (d *Disk) loadKeysFromDisk() error {
	data, err := os.ReadFile(d.keysFilePath)
	if err != nil {
		return nil // No keys file exists, no error
	}

	keys := make(map[string]string)
	if err := json.Unmarshal(data, &keys); err != nil {
		return fmt.Errorf("failed to unmarshal keys from disk: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	for id, keyBytes := range keys {
		signer, err := convertToSigner([]byte(keyBytes))
		if err != nil {
			return fmt.Errorf("failed to create key entry: %w", err)
		}

		d.entries[id] = &KeyEntry{
			PrivateKey: signer,
			PublicKey:  signer.Public(),
			id:         id,
		}
	}

	return nil
}

// SaveKeysToDisk saves the keys in the key manager to disk.
func (d *Disk) saveKeysToDisk() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	keys := make(map[string]string)
	for id, entry := range d.entries {
		// Encode private key as PEM block
		keyBytes, err := x509.MarshalPKCS8PrivateKey(entry.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to marshal private key: %w", err)
		}

		// Encode PEM block as string
		keyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyBytes,
		})
		keys[id] = string(keyPEM)
	}

	data, err := json.Marshal(keys)
	if err != nil {
		return fmt.Errorf("failed to serialize keys: %w", err)
	}

	if err := os.WriteFile(d.keysFilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write keys to disk: %w", err)
	}

	return nil
}

func convertToSigner(keyBytes []byte) (crypto.Signer, error) {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	switch key := key.(type) {
	case *rsa.PrivateKey:
		return key, nil
	case *ecdsa.PrivateKey:
		return key, nil
	default:
		return nil, errors.New("unsupported private key type")
	}
}
