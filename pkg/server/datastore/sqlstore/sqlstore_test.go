package sqlstore

import (
	"context"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/stretchr/testify/assert"
)

var (
	ctx = context.Background()
	ds  = SQLStore{}
)

func TestCreateDB(t *testing.T) {
	assert.Nil(t, ds.OpenDB("gorm_open.db", "sqlite3"))
}

func TestCreateMember(t *testing.T) {
	assert.Nil(t, ds.OpenDB("gorm_open.db", "sqlite3"))
	ent := entity.Member{
		Description: "Test Mem",
		TrustDomain: "test.org",
	}
	_, err := ds.CreateMember(ctx, &ent)
	assert.Nil(t, err)
	assert.Equal(t, ent.Description, "Test Mem")
	assert.Equal(t, ent.TrustDomain, "test.org")

}
