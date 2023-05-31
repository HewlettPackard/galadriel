package fakekeymanager

import (
	"context"
	"crypto"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/keymanager"
)

// KeyManager is a fake key manager that returns the same key for all requests
type KeyManager struct {
	Key crypto.Signer
}

func (f KeyManager) GenerateKey(ctx context.Context, id string, keyType cryptoutil.KeyType) (keymanager.Key, error) {
	return &keymanager.KeyEntry{
		PrivateKey: f.Key,
		PublicKey:  f.Key.Public(),
	}, nil
}

func (f KeyManager) GetKey(ctx context.Context, id string) (keymanager.Key, error) {
	return &keymanager.KeyEntry{
		PrivateKey: f.Key,
		PublicKey:  f.Key.Public(),
	}, nil
}

func (f KeyManager) GetKeys(ctx context.Context) ([]keymanager.Key, error) {
	return []keymanager.Key{&keymanager.KeyEntry{
		PrivateKey: f.Key,
		PublicKey:  f.Key.Public(),
	}}, nil
}
