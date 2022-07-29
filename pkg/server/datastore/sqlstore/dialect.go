package sqlstore

import "gorm.io/gorm"

type dialect interface {
	connect(ConnectionString string) (db *gorm.DB, err error)
}
