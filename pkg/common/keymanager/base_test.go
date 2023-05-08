package keymanager

import (
	"context"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSetsConfigDefaults(t *testing.T) {
	b := New(&Config{})
	assert.Equal(t, &defaultGenerator{}, b.generator)
}

func TestGenerateKey(t *testing.T) {
	b := New(&Config{})
	ctx := context.Background()

	key, err := b.GenerateKey(ctx, "foo", cryptoutil.RSA2048)
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, "foo", key.ID())
	assert.Equal(t, key, b.entries["foo"])

	key, err = b.GenerateKey(ctx, "bar", cryptoutil.RSA4096)
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, "bar", key.ID())
	assert.Equal(t, key, b.entries["bar"])
}

func TestGenerateKeyOverridesKey(t *testing.T) {
	b := New(&Config{})
	ctx := context.Background()

	key, err := b.GenerateKey(ctx, "foo", cryptoutil.RSA2048)
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, "foo", key.ID())
	assert.Equal(t, key, b.entries["foo"])

	key, err = b.GenerateKey(ctx, "foo", cryptoutil.RSA4096)
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, "foo", key.ID())
	assert.Equal(t, key, b.entries["foo"])
}

func TestGenerateKeyFailsWithInvalidKeyType(t *testing.T) {
	b := New(&Config{})
	ctx := context.Background()

	key, err := b.GenerateKey(ctx, "foo", cryptoutil.RSA2048)
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, key, b.entries["foo"])

	key, err = b.GenerateKey(ctx, "foo", cryptoutil.RSA4096)
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, key, b.entries["foo"])
}

func TestGenerateKeyFailsWithEmptyKeyID(t *testing.T) {
	b := New(&Config{})
	ctx := context.Background()

	key, err := b.GenerateKey(ctx, "", cryptoutil.RSA2048)
	assert.Error(t, err)
	assert.Nil(t, key)
}

func TestGetKey(t *testing.T) {
	b := New(&Config{})
	ctx := context.Background()

	key, err := b.GenerateKey(ctx, "foo", cryptoutil.RSA2048)
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, key, b.entries["foo"])

	key, err = b.GetKey(ctx, "foo")
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, key, b.entries["foo"])
}

func TestGetKeyFailsWithEmptyKeyID(t *testing.T) {
	b := New(&Config{})
	ctx := context.Background()

	key, err := b.GetKey(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, key)
}

func TestGetKeyFailsWithUnknownKeyID(t *testing.T) {
	b := New(&Config{})
	ctx := context.Background()

	key, err := b.GetKey(ctx, "foo")
	assert.Error(t, err)
	assert.Nil(t, key)
}

func TestGetKeys(t *testing.T) {
	b := New(&Config{})
	ctx := context.Background()

	key1, err := b.GenerateKey(ctx, "foo", cryptoutil.RSA2048)
	assert.NoError(t, err)
	assert.NotNil(t, key1)

	key2, err := b.GenerateKey(ctx, "bar", cryptoutil.RSA2048)
	assert.NoError(t, err)
	assert.NotNil(t, key2)

	keys, err := b.GetKeys(ctx)
	assert.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Equal(t, key1, b.entries["foo"])
	assert.Equal(t, key2, b.entries["bar"])
}

func TestGetKeysEmptyKeyManager(t *testing.T) {
	b := New(&Config{})
	ctx := context.Background()

	keys, err := b.GetKeys(ctx)
	assert.NoError(t, err)
	assert.Len(t, keys, 0)
}
