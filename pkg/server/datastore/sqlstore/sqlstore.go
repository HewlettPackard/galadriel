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
func (ds *SQLStore) CreateRelationship(ctx context.Context, relationship *entity.Relationship, sourceMemberID uuid.UUID, targetMemberID uuid.UUID) (*entity.Relationship, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) RetrieveMemberByID(ctx context.Context, memberID uuid.UUID) (*entity.Member, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) RetrieveMembershipByID(ctx context.Context, membershipID uuid.UUID) (*entity.Membership, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) RetrieveRelationshipBySourceandTargetID(ctx context.Context, sourceMemberID uuid.UUID, targetMemberID uuid.UUID) (*entity.Relationship, error) {
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
func (ds *SQLStore) UpdateTrust(ctx context.Context, trustbundle *common.Bundle) (*common.Bundle, error) {
	return nil, fmt.Errorf("not implemented")
}
func (ds *SQLStore) DeleteMemberbyID(ctx context.Context, memberID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
func (ds *SQLStore) DeleteMembershipsByID(ctx context.Context, memberid uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
func (ds *SQLStore) DeleteRelationshipsByID(ctx context.Context, memberid uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
func (ds *SQLStore) DeleteTrustbundlesByID(ctx context.Context, memberid uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
func (ds *SQLStore) DeleteMembershipByID(ctx context.Context, membershipID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
func (ds *SQLStore) DeleteRelationshipBySourceTargetID(ctx context.Context, sourceMemberID uuid.UUID, targetMemberID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
func (ds *SQLStore) DeleteTrustBundleByID(ctx context.Context, memberID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
