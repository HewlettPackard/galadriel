package sqlstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	// SQLite database type
	SQLite = "sqlite3"
)

var _ datastore.DataStore = &SQLStore{}

type SQLStore struct {
	db *gorm.DB
}

// OpenDB Opens a database connection by connection string and DBtype
func (ds *SQLStore) OpenDB(connectionString, dbtype string) (err error) {
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

// CreateMember implements the CreateMember function from Datastore
// Creates a new Member in the database. Returns error if fails.
func (ds *SQLStore) CreateMember(_ context.Context, member *entity.Member) (*entity.Member, error) {
	return nil, errors.New("not implemented")
}

func (ds *SQLStore) CreateMembership(ctx context.Context, membership *entity.Membership) (*entity.Membership, error) {
	return nil, errors.New("not implemented")
}

func (ds *SQLStore) CreateRelationship(ctx context.Context, relationship *entity.Relationship) (*entity.Relationship, error) {
	return nil, errors.New("not implemented")
}

func (ds *SQLStore) GetMemberByID(ctx context.Context, memberID uuid.UUID) (*entity.Member, error) {
	return nil, errors.New("not implemented")
}
