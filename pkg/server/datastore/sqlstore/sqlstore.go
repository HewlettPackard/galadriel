package sqlstore

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
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
func (ds *Plugin) CreateMember(ctx echo.Context, entitymember *entity.Member) (*entity.Member, error) {

	dbmember := Member{
		Description: (*entitymember).Description,
		TrustDomain: (*entitymember).TrustDomain,
	}

	if err := ds.db.Where(&dbmember).FirstOrCreate(&dbmember).Error; err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	entitymember.ID = dbmember.ID
	return entitymember, nil
}
