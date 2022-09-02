package sqlstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitDB(t *testing.T) {
	dialectvar := sqliteDB{}
	db, err := dialectvar.connect("gorm_init.db")
	assert.Nil(t, err)
	assert.Nil(t, initDB(db))
}

func TestMigrateDB(t *testing.T) {
	dialectvar := sqliteDB{}
	db, err := dialectvar.connect("gorm_migrate.db")
	assert.Nil(t, err)
	assert.Nil(t, migrateDB(db))
}
