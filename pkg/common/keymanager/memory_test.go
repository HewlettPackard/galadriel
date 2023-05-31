package keymanager

import (
	"context"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	km := newBase(nil)
	assert.NotNil(t, km)
}

func TestKeyManager(t *testing.T) {
	km := newBase(nil)
	ctx := context.Background()

	key1, err := km.GenerateKey(ctx, "foo", cryptoutil.RSA2048)
	assert.NoError(t, err)
	assert.NotNil(t, key1)
	assert.Equal(t, "foo", key1.ID())

	key2, err := km.GenerateKey(ctx, "bar", cryptoutil.RSA2048)
	assert.NoError(t, err)
	assert.NotNil(t, key2)
	assert.Equal(t, "bar", key2.ID())

	key, err := km.GetKey(ctx, "foo")
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, key1, key)

	key, err = km.GetKey(ctx, "bar")
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, key2, key)

	keys, err := km.GetKeys(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.Equal(t, 2, len(keys))
	assert.Contains(t, keys, key1)
	assert.Contains(t, keys, key2)
}
