package sqlstore

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type sqliteDB struct{}

func (sqliteDB) connect(connectionString string) (db *gorm.DB, err error) {
	db, err = gorm.Open(sqlite.Open(ConnectionString), &gorm.Config{})
	if err != nil {

		return nil, err
	}
	return db, nil
}
