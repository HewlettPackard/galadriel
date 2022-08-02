package sqlstore

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type postgresDB struct{}

func (postgresDB) connect(connectionString string) (db *gorm.DB, err error) {
	db, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {

		return nil, err
	}
	return db, nil
}
