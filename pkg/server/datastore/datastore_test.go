package datastore_test

import (
	"context"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJoinTokenMethods(t *testing.T) {

	d := datastore.NewMemStore()

	tokenStr := "token"
	token := datastore.AccessToken{
		Token:  tokenStr,
		Expiry: time.Now(),
	}

	ctx := context.Background()

	err := d.CreateAccessToken(ctx, &token)
	if err != nil {
		t.Error(err)
	}

	storedToken, err := d.FetchJoinToken(ctx, tokenStr)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, tokenStr, storedToken.Token)

	err = d.DeleteJoinToken(ctx, tokenStr)
	if err != nil {
		t.Error(err)
	}

	_, err = d.FetchJoinToken(ctx, tokenStr)
	assert.Equal(t, "token not found", err.Error())

	err = d.DeleteJoinToken(ctx, tokenStr)
	assert.Equal(t, "token not found", err.Error())
}
