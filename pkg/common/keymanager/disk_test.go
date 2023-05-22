package keymanager

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiskKeyManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "keymanager_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dataDir := filepath.Join(tempDir, "keys-test.json")

	// Create a new Disk key manager
	keyManager, err := NewDiskKeyManager(nil, dataDir)
	require.NoError(t, err)

	// Generate a new key pair
	keyID := "key1"
	keyType := cryptoutil.RSA2048
	key, err := keyManager.GenerateKey(context.Background(), keyID, keyType)
	require.NoError(t, err)
	require.NotNil(t, key)

	// Get the generated key
	gotKey, err := keyManager.GetKey(context.Background(), keyID)
	require.NoError(t, err)
	require.NotNil(t, gotKey)

	// Verify the generated key's ID and type
	assert.Equal(t, keyID, gotKey.ID())

	// Verify the generated key's
	assert.NotNil(t, gotKey.Signer())

	// Load the Disk key manager from disk
	loadedKeyManager, err := NewDiskKeyManager(nil, dataDir)
	require.NoError(t, err)

	// Get the loaded key
	loadedKey, err := loadedKeyManager.GetKey(context.Background(), keyID)
	require.NoError(t, err)
	require.NotNil(t, loadedKey)

	// Verify the loaded key's ID and type
	assert.Equal(t, keyID, loadedKey.ID())

	// Verify the loaded key's
	assert.NotNil(t, loadedKey.Signer())
}
