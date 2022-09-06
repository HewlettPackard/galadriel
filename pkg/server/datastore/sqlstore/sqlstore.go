package sqlstore

import (
	"context"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/google/uuid"
	"github.com/spiffe/spire/proto/spire/common"
	"gorm.io/gorm"
)

const (
	// SQLite database type
	SQLite = "sqlite3"
)

// Implementation compliance test
var _ datastore.DataStore = &SQLStore{}

type SQLStore struct {
	db *gorm.DB
}

// Open a database connection by connection string and DBtype
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

// Implements the CreateMember function from Datastore
// Creates a new Member in the database. Returns error if fails.
func (ds *SQLStore) CreateMember(ctx context.Context, entitymember *entity.Member) (*entity.Member, error) {
	var err error
	dbmember := Member{
		Description: (*entitymember).Description,
		TrustDomain: (*entitymember).TrustDomain,
	}

	if err = ds.db.Where(&dbmember).FirstOrCreate(&dbmember).Error; err != nil {
		return nil, fmt.Errorf("sqlstore error: %v", err)
	}
	entitymember.ID = dbmember.ID
	entitymember.CreatedAt = dbmember.CreatedAt
	entitymember.UpdatedAt = dbmember.UpdatedAt
	return entitymember, nil
}

func (ds *SQLStore) CreateMembership(ctx context.Context, membership *entity.Membership) (*entity.Membership, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) CreateTrustBundle(ctx context.Context, trustBundle *common.Bundle, memberID uuid.UUID) (*common.Bundle, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) CreateRelationship(ctx context.Context, relationship *entity.Relationship) (*entity.Relationship, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) RetrieveMemberByID(ctx context.Context, memberID uuid.UUID) (*entity.Member, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) RetrieveMembershipByID(ctx context.Context, membershipID uuid.UUID) (*entity.Membership, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) RetrieveRelationshipByID(ctx context.Context, relationshipID uuid.UUID) (*entity.Relationship, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) RetrieveTrustBundleByID(ctx context.Context, trustID uuid.UUID) (*common.Bundle, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) UpdateMember(ctx context.Context, member *entity.Member) (*entity.Member, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) UpdateMembership(ctx context.Context, membership *entity.Membership) (*entity.Membership, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) UpdateTrustBundle(ctx context.Context, trustbundle *common.Bundle) (*common.Bundle, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) UpdateRelationship(ctx context.Context, relationship *entity.Relationship) (*common.Bundle, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) DeleteMemberByID(ctx context.Context, memberID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
func (ds *SQLStore) DeleteMembershipByID(ctx context.Context, membershipID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
func (ds *SQLStore) DeleteRelationshipByID(ctx context.Context, relationshipID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
func (ds *SQLStore) DeleteTrustBundleByID(ctx context.Context, memberID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
