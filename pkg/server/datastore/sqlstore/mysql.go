package sqlstore

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type mysqlDB struct{}

func (mysqlDB) connect(connectionString string) (db *gorm.DB, err error) {
	db, err = gorm.Open(mysql.Open(ConnectionString), &gorm.Config{})
	if err != nil {

		return nil, err
	}
	return db, nil
}
