package sqlstore

import (
	"fmt"
	"net/url"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type sqliteDB struct{}

func (s sqliteDB) connect(connectionString string) (db *gorm.DB, err error) {
	if connectionString, err = s.addQueryValues(connectionString); err != nil {
		return nil, err
	}
	db, err = gorm.Open(sqlite.Open(connectionString), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// The following function follows the Spire embellish function
// and adds query values supported by github.com/mattn/go-sqlite3
// to enable journal mode and foreign key support.
func (sqliteDB) addQueryValues(connectionString string) (string, error) {
	u, err := url.Parse(connectionString)
	if err != nil {
		return "", err
	}

	switch {
	case u.Scheme == "":
		u.Scheme = "file"
		u.Opaque, u.Path = u.Path, ""
	case u.Scheme != "file":
		return "", fmt.Errorf("unsupported scheme %q", u.Scheme)
	}

	q := u.Query()
	q.Set("_foreign_keys", "ON")
	q.Set("_journal_mode", "WAL")
	u.RawQuery = q.Encode()
	return u.String(), nil
}
