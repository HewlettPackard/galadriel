package sqlstore

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/server/api/management"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

const (
	// SQLite database type
	SQLite = "sqlite3"
)

type Plugin struct {
	db *gorm.DB
}

// Open a database connection by connection string and DBtype
func (ds *Plugin) OpenDB(connectionString, dbtype string) (err error) {
	var dialectvar dialect

	switch dbtype {
	case SQLite:
		dialectvar = sqliteDB{}
	default:
		return fmt.Errorf("unsupported database_type: %s" + dbtype)
	}

	if ds.db, err = dialectvar.connect(connectionString); err != nil {
		return fmt.Errorf("error connecting to: %s", connectionString)
	}
	return migrateDB(ds.db)
}

// Implements the CreateMember function from Datastore
// Creates a new Member in the database. Returns error if fails.
func (ds *Plugin) CreateMember(ctx echo.Context, server *management.SpireServer) (*management.SpireServer, error) {

	member := Member{Description: (*server).Description}

	if err := ds.db.Where(&member).FirstOrCreate(&member).Error; err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	member.ID = uint(server.Id)
	return server, nil
}
