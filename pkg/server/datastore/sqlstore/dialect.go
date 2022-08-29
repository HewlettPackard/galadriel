package sqlstore

import "gorm.io/gorm"

// Dialect inteface for support for multiple DB types
type dialect interface {
	connect(connectionString string) (db *gorm.DB, err error)
}
