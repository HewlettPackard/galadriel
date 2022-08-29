package sqlstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	ds = Plugin{}
)

func TestCreateDB(t *testing.T) {
	assert.Nil(t, ds.OpenDB("gorm_org.db", "sqlite3"))
}
