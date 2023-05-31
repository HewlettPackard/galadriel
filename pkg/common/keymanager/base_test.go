package keymanager

import (
	"context"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/stretchr/testify/assert"
)

func setup() (*base, context.Context) {
	b := newBase(&Config{})
	ctx := context.Background()
	return b, ctx
}

func TestNewSetsConfigDefaults(t *testing.T) {
	b, _ := setup()
	assert.Equal(t, &defaultGenerator{}, b.generator)
}

func TestGenerateKey(t *testing.T) {
	b, ctx := setup()

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
	b, ctx := setup()

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
	b, ctx := setup()

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
	b, ctx := setup()

	key, err := b.GenerateKey(ctx, "", cryptoutil.RSA2048)
	assert.Error(t, err)
	assert.Nil(t, key)
}

func TestGetKey(t *testing.T) {
	b, ctx := setup()

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
	b, ctx := setup()

	key, err := b.GetKey(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, key)
}

func TestGetKeyFailsWithUnknownKeyID(t *testing.T) {
	b, ctx := setup()

	key, err := b.GetKey(ctx, "foo")
	assert.Error(t, err)
	assert.Nil(t, key)
}
