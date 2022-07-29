package sqlstore

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type mysqlDB struct{}

func (my mysqlDB) connect(ConnectionString string) (db *gorm.DB, err error) {
	db, err = gorm.Open(mysql.Open(ConnectionString), &gorm.Config{})
	if err != nil {

		return nil, err
	}
	return db, nil
}
