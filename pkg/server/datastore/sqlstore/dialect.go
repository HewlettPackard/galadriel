package sqlstore

import "gorm.io/gorm"

type dialect interface {
	connect(connectionString string) (db *gorm.DB, err error)
}
