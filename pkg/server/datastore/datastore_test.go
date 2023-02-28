package datastore_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	postgresImage = "15-alpine"
	user          = "test_user"
	password      = "test_password"
	dbname        = "test_db"
)

var (
	spiffeTD1 = spiffeid.RequireTrustDomainFromString("foo.test")
	spiffeTD2 = spiffeid.RequireTrustDomainFromString("bar.test")
	spiffeTD3 = spiffeid.RequireTrustDomainFromString("baz.test")
)

func TestCreateTrustDomain(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	req1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1 := createTrustDomain(t, ctx, ds, req1)
	assert.Equal(t, req1.Name, td1.Name)
	require.NotNil(t, td1.ID)
	require.NotNil(t, td1.CreatedAt)
	require.NotNil(t, td1.UpdatedAt)

	// Look up trust domain stored in DB and compare
	stored, err := ds.FindTrustDomainByID(ctx, td1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, td1, stored)
}

func TestUpdateTrustDomain(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	req1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1 := createTrustDomain(t, ctx, ds, req1)

	td1.Description = "updated_description"
	td1.HarvesterSpiffeID = spiffeid.RequireFromString("spiffe://domain/test")
	td1.OnboardingBundle = []byte{1, 2, 3}

	// Update Trust Domain
	updated1, err := ds.CreateOrUpdateTrustDomain(ctx, td1)
	require.NoError(t, err)
	require.NotNil(t, updated1)

	// Look up trust domain stored in DB and compare
	stored, err := ds.FindTrustDomainByID(ctx, td1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, td1.ID, stored.ID)
	assert.Equal(t, td1.Description, stored.Description)
	assert.Equal(t, td1.HarvesterSpiffeID, stored.HarvesterSpiffeID)
	assert.Equal(t, td1.OnboardingBundle, stored.OnboardingBundle)
}

func TestTrustFindDomainByName(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	req1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1 := createTrustDomain(t, ctx, ds, req1)

	req2 := &entity.TrustDomain{
		Name: spiffeTD2,
	}
	td2 := createTrustDomain(t, ctx, ds, req2)

	stored1, err := ds.FindTrustDomainByName(ctx, td1.Name)
	require.NoError(t, err)
	assert.Equal(t, td1, stored1)

	stored2, err := ds.FindTrustDomainByName(ctx, td2.Name)
	require.NoError(t, err)
	assert.Equal(t, td2, stored2)
}

func TestDeleteTrustDomain(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	req1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1 := createTrustDomain(t, ctx, ds, req1)

	req2 := &entity.TrustDomain{
		Name: spiffeTD2,
	}
	td2 := createTrustDomain(t, ctx, ds, req2)

	err := ds.DeleteTrustDomain(ctx, td1.ID.UUID)
	require.NoError(t, err)

	stored, err := ds.FindTrustDomainByID(ctx, td1.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	err = ds.DeleteTrustDomain(ctx, td2.ID.UUID)
	require.NoError(t, err)

	stored, err = ds.FindTrustDomainByID(ctx, td2.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)
}

func TestListTrustDomains(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	req1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1 := createTrustDomain(t, ctx, ds, req1)

	req2 := &entity.TrustDomain{
		Name: spiffeTD2,
	}
	td2 := createTrustDomain(t, ctx, ds, req2)

	list, err := ds.ListTrustDomains(ctx)
	require.NoError(t, err)
	require.NotNil(t, list)
	assert.Equal(t, 2, len(list))
	assert.Contains(t, list, td1)
	assert.Contains(t, list, td2)
}

func TestTrustDomainUniqueTrustDomainConstraint(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	td1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	_, err = datastore.CreateOrUpdateTrustDomain(ctx, td1)
	require.NoError(t, err)

	// second trustDomain with same trust domain
	td2 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	_, err = datastore.CreateOrUpdateTrustDomain(ctx, td2)
	require.Error(t, err)

	wrappedErr := errors.Unwrap(err)
	errCode := wrappedErr.(*pgconn.PgError).SQLState()
	assert.Equal(t, pgerrcode.UniqueViolation, errCode, "Unique constraint violation was expected")
}

func TestCreateRelationship(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	// Setup
	// Create TrustDomains
	td1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1 = createTrustDomain(t, ctx, ds, td1)

	td2 := &entity.TrustDomain{
		Name: spiffeTD2,
	}
	td2 = createTrustDomain(t, ctx, ds, td2)

	// Create relationship TrustDomain1 -- TrustDomain2
	req1 := &entity.Relationship{
		TrustDomainAID: td1.ID.UUID,
		TrustDomainBID: td2.ID.UUID,
	}

	relationship1 := createRelationship(t, ctx, ds, req1)
	require.NotNil(t, relationship1.ID)
	require.NotNil(t, relationship1.CreatedAt)
	require.NotNil(t, relationship1.UpdatedAt)
	assert.Equal(t, req1.TrustDomainAID, relationship1.TrustDomainAID)
	assert.Equal(t, req1.TrustDomainBID, relationship1.TrustDomainBID)

	// Look up relationship in DB and compare
	stored, err := ds.FindRelationshipByID(ctx, relationship1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, relationship1, stored)
}

func TestUpdateRelationship(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	// Setup
	// Create TrustDomains
	td1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1 = createTrustDomain(t, ctx, ds, td1)

	td2 := &entity.TrustDomain{
		Name: spiffeTD2,
	}
	td2 = createTrustDomain(t, ctx, ds, td2)

	// Create relationship TrustDomain1 -- TrustDomain2
	req1 := &entity.Relationship{
		TrustDomainAID: td1.ID.UUID,
		TrustDomainBID: td2.ID.UUID,
	}

	relationship1 := createRelationship(t, ctx, ds, req1)

	relationship1.TrustDomainAConsent = true
	relationship1.TrustDomainBConsent = true

	updated1, err := ds.CreateOrUpdateRelationship(ctx, relationship1)
	require.NoError(t, err)

	// Look up relationship in DB and compare
	stored, err := ds.FindRelationshipByID(ctx, updated1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, updated1, stored)
	assert.True(t, stored.TrustDomainAConsent)
	assert.True(t, stored.TrustDomainBConsent)
}

func TestFindRelationshipByTrustDomain(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	// Setup
	// Create TrustDomains
	td1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1 = createTrustDomain(t, ctx, ds, td1)

	td2 := &entity.TrustDomain{
		Name: spiffeTD2,
	}
	td2 = createTrustDomain(t, ctx, ds, td2)

	td3 := &entity.TrustDomain{
		Name: spiffeTD3,
	}
	td3 = createTrustDomain(t, ctx, ds, td3)

	// Create relationship TrustDomain1 -- TrustDomain2
	req1 := &entity.Relationship{
		TrustDomainAID: td1.ID.UUID,
		TrustDomainBID: td2.ID.UUID,
	}

	relationship1 := createRelationship(t, ctx, ds, req1)

	// Create relationship TrustDomain1 -- TrustDomain3
	req2 := &entity.Relationship{
		TrustDomainAID: td1.ID.UUID,
		TrustDomainBID: td3.ID.UUID,
	}
	relationship2 := createRelationship(t, ctx, ds, req2)

	// Find relationships by TrustDomainID
	relationships, err := ds.FindRelationshipsByTrustDomainID(ctx, td1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(relationships))
	assert.Contains(t, relationships, relationship1)
	assert.Contains(t, relationships, relationship2)

	relationships, err = ds.FindRelationshipsByTrustDomainID(ctx, td2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(relationships))
	assert.Contains(t, relationships, relationship1)
}

func TestDeleteRelationship(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	// Setup
	// Create TrustDomains
	td1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1 = createTrustDomain(t, ctx, ds, td1)

	td2 := &entity.TrustDomain{
		Name: spiffeTD2,
	}
	td2 = createTrustDomain(t, ctx, ds, td2)

	td3 := &entity.TrustDomain{
		Name: spiffeTD3,
	}
	td3 = createTrustDomain(t, ctx, ds, td3)

	// Create relationship TrustDomain1 -- TrustDomain2
	req1 := &entity.Relationship{
		TrustDomainAID: td1.ID.UUID,
		TrustDomainBID: td2.ID.UUID,
	}

	relationship1 := createRelationship(t, ctx, ds, req1)

	// Create relationship TrustDomain1 -- TrustDomain3
	req2 := &entity.Relationship{
		TrustDomainAID: td1.ID.UUID,
		TrustDomainBID: td3.ID.UUID,
	}
	relationship2 := createRelationship(t, ctx, ds, req2)

	// Delete relationships
	err := ds.DeleteRelationship(ctx, relationship1.ID.UUID)
	require.NoError(t, err)
	stored, err := ds.FindRelationshipByID(ctx, relationship1.ID.UUID)
	require.NoError(t, err)
	assert.Nil(t, stored)

	err = ds.DeleteRelationship(ctx, relationship2.ID.UUID)
	require.NoError(t, err)
	stored, err = ds.FindRelationshipByID(ctx, relationship2.ID.UUID)
	require.NoError(t, err)
	assert.Nil(t, stored)
}

func TestListRelationships(t *testing.T) {
	t.Parallel()
	ds, ctx := setupTest(t)

	// Setup
	// Create TrustDomains
	td1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1 = createTrustDomain(t, ctx, ds, td1)

	td2 := &entity.TrustDomain{
		Name: spiffeTD2,
	}
	td2 = createTrustDomain(t, ctx, ds, td2)

	td3 := &entity.TrustDomain{
		Name: spiffeTD3,
	}
	td3 = createTrustDomain(t, ctx, ds, td3)

	// Create relationship TrustDomain1 -- TrustDomain2
	req1 := &entity.Relationship{
		TrustDomainAID: td1.ID.UUID,
		TrustDomainBID: td2.ID.UUID,
	}

	relationship1 := createRelationship(t, ctx, ds, req1)

	// Create relationship TrustDomain1 -- TrustDomain3
	req2 := &entity.Relationship{
		TrustDomainAID: td1.ID.UUID,
		TrustDomainBID: td3.ID.UUID,
	}
	relationship2 := createRelationship(t, ctx, ds, req2)

	relationships, err := ds.ListRelationships(ctx)
	require.NoError(t, err)
	assert.Contains(t, relationships, relationship1)
	assert.Contains(t, relationships, relationship2)
}

func TestRelationshipForeignKeysConstraints(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	td1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1, err = datastore.CreateOrUpdateTrustDomain(ctx, td1)
	require.NoError(t, err)

	td2 := &entity.TrustDomain{
		Name: spiffeTD2,
	}
	td2, err = datastore.CreateOrUpdateTrustDomain(ctx, td2)
	require.NoError(t, err)

	relationship1 := &entity.Relationship{
		TrustDomainAID: td1.ID.UUID,
		TrustDomainBID: td2.ID.UUID,
	}
	relationship1, err = datastore.CreateOrUpdateRelationship(ctx, relationship1)
	require.NoError(t, err)

	// Cannot add a new relationship for the same TrustDomains
	relationship1.ID = uuid.NullUUID{}
	_, err = datastore.CreateOrUpdateRelationship(ctx, relationship1)
	require.Error(t, err)

	wrappedErr := errors.Unwrap(err)
	errCode := wrappedErr.(*pgconn.PgError).SQLState()
	assert.Equal(t, pgerrcode.UniqueViolation, errCode, "Unique constraint violation error was expected")

	// Cannot delete Trust Domain that has a relationship associated
	err = datastore.DeleteTrustDomain(ctx, td1.ID.UUID)
	require.Error(t, err)

	wrappedErr = errors.Unwrap(err)
	errCode = wrappedErr.(*pgconn.PgError).SQLState()
	assert.Equal(t, pgerrcode.ForeignKeyViolation, errCode, "Foreign key violation error was expected")

	// Cannot delete Trust Domain that has a relationship associated
	err = datastore.DeleteTrustDomain(ctx, td2.ID.UUID)
	require.Error(t, err)

	wrappedErr = errors.Unwrap(err)
	errCode = wrappedErr.(*pgconn.PgError).SQLState()
	assert.Equal(t, pgerrcode.ForeignKeyViolation, errCode, "Foreign key violation error was expected")
}

func createTrustDomain(t *testing.T, ctx context.Context, ds *datastore.SQLDatastore, req *entity.TrustDomain) *entity.TrustDomain {
	td1, err := ds.CreateOrUpdateTrustDomain(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, td1.ID)

	return td1
}

func createRelationship(t *testing.T, ctx context.Context, ds *datastore.SQLDatastore, req *entity.Relationship) *entity.Relationship {
	td1, err := ds.CreateOrUpdateRelationship(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, td1.ID)

	return td1
}

func TestCRUDBundle(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Create trustDomains to associate the bundles
	td1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1, err = datastore.CreateOrUpdateTrustDomain(ctx, td1)
	require.NoError(t, err)
	require.NotNil(t, td1.ID)

	td2 := &entity.TrustDomain{
		Name: spiffeTD2,
	}
	td2, err = datastore.CreateOrUpdateTrustDomain(ctx, td2)
	require.NoError(t, err)
	require.NotNil(t, td2.ID)

	// Create first Data - trustDomain-1
	req1 := &entity.Bundle{
		Data:               []byte{1, 2, 3},
		Digest:             []byte{10, 20, 30},
		Signature:          []byte{4, 2},
		DigestAlgorithm:    "MD5",
		SignatureAlgorithm: "RSA",
		SigningCert:        []byte{50, 60},
		TrustDomainID:      td1.ID.UUID,
	}

	b1, err := datastore.CreateOrUpdateBundle(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, b1)
	assert.Equal(t, req1.Data, b1.Data)
	assert.Equal(t, req1.Digest, b1.Digest)
	assert.Equal(t, req1.Signature, b1.Signature)
	assert.Equal(t, req1.DigestAlgorithm, b1.DigestAlgorithm)
	assert.Equal(t, req1.SignatureAlgorithm, b1.SignatureAlgorithm)
	assert.Equal(t, req1.SigningCert, b1.SigningCert)
	assert.Equal(t, req1.TrustDomainID, b1.TrustDomainID)

	// Look up bundle stored in DB and compare
	stored, err := datastore.FindBundleByID(ctx, b1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, b1, stored)

	// Create second Data -> trustDomain-2
	req2 := &entity.Bundle{
		Data:               []byte{10, 20, 30},
		Digest:             []byte{100, 200, 30},
		Signature:          []byte{40, 20},
		DigestAlgorithm:    "MD6",
		SignatureAlgorithm: "ECDSA",
		SigningCert:        []byte{80, 90},
		TrustDomainID:      td2.ID.UUID,
	}

	b2, err := datastore.CreateOrUpdateBundle(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, b1)
	assert.Equal(t, req2.Data, b2.Data)
	assert.Equal(t, req2.Digest, b2.Digest)
	assert.Equal(t, req2.Signature, b2.Signature)
	assert.Equal(t, req2.DigestAlgorithm, b2.DigestAlgorithm)
	assert.Equal(t, req2.SignatureAlgorithm, b2.SignatureAlgorithm)
	assert.Equal(t, req2.SigningCert, b2.SigningCert)
	assert.Equal(t, req2.TrustDomainID, b2.TrustDomainID)

	// Look up bundle stored in DB and compare
	stored, err = datastore.FindBundleByID(ctx, b2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, b2, stored)

	// Find bundles by TrustDomainID
	stored, err = datastore.FindBundleByTrustDomainID(ctx, td1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, b1, stored)

	stored, err = datastore.FindBundleByTrustDomainID(ctx, td2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, b2, stored)

	// Update Data
	b1.Data = []byte{'a', 'b', 'c'}
	b1.Digest = []byte{'c', 'd', 'e'}
	b1.Signature = []byte{'f', 'g', 'h'}
	b1.DigestAlgorithm = "other"
	b1.SignatureAlgorithm = "other-alg"
	b1.SigningCert = []byte{'f', 'g', 'h'}

	updated, err := datastore.CreateOrUpdateBundle(ctx, b1)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, b1.Data, updated.Data)
	assert.Equal(t, b1.Digest, updated.Digest)
	assert.Equal(t, b1.Signature, updated.Signature)
	assert.Equal(t, b1.DigestAlgorithm, updated.DigestAlgorithm)
	assert.Equal(t, b1.SignatureAlgorithm, updated.SignatureAlgorithm)
	assert.Equal(t, b1.SigningCert, updated.SigningCert)
	assert.Equal(t, b1.TrustDomainID, updated.TrustDomainID)

	// Look up bundle stored in DB and compare
	stored, err = datastore.FindBundleByID(ctx, b1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, updated, stored)

	// List bundles
	bundles, err := datastore.ListBundles(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(bundles))
	require.Contains(t, bundles, updated)
	require.Contains(t, bundles, b2)

	// Delete first bundle
	err = datastore.DeleteBundle(ctx, b1.ID.UUID)
	require.NoError(t, err)

	// Look up deleted bundle
	stored, err = datastore.FindBundleByID(ctx, b1.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	// Delete second bundle
	err = datastore.DeleteBundle(ctx, b2.ID.UUID)
	require.NoError(t, err)

	// Look up deleted bundle
	stored, err = datastore.FindBundleByID(ctx, b2.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)
}

func TestBundleUniqueTrustDomainConstraint(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Create trustDomain to associate the bundles
	td1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1, err = datastore.CreateOrUpdateTrustDomain(ctx, td1)
	require.NoError(t, err)
	require.NotNil(t, td1.ID)

	// Create Data
	b1 := &entity.Bundle{
		Data:               []byte{1, 2, 3},
		Digest:             []byte{10, 20, 30},
		Signature:          []byte{4, 2},
		DigestAlgorithm:    "MD5",
		SignatureAlgorithm: "RSA",
		SigningCert:        []byte{50, 60},
		TrustDomainID:      td1.ID.UUID,
	}
	b1, err = datastore.CreateOrUpdateBundle(ctx, b1)
	require.NoError(t, err)
	require.NotNil(t, b1)

	// Create second Data associated to same trustDomain
	b2 := &entity.Bundle{
		Data:               []byte{10, 20, 30},
		Digest:             []byte{100, 200, 30},
		Signature:          []byte{40, 20},
		DigestAlgorithm:    "MD6",
		SignatureAlgorithm: "ECDSA",
		SigningCert:        []byte{80, 90},
		TrustDomainID:      td1.ID.UUID,
	}
	b2, err = datastore.CreateOrUpdateBundle(ctx, b2)
	require.Error(t, err)
	require.Nil(t, b2)

	wrappedErr := errors.Unwrap(err)
	errCode := wrappedErr.(*pgconn.PgError).SQLState()
	assert.Equal(t, pgerrcode.UniqueViolation, errCode, "Unique constraint violation error was expected")
}

func TestCRUDJoinToken(t *testing.T) {
	t.Parallel()
	datastore, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the datastore")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	//// Create trustDomains to associate the join tokens
	td1 := &entity.TrustDomain{
		Name: spiffeTD1,
	}
	td1, err = datastore.CreateOrUpdateTrustDomain(ctx, td1)
	require.NoError(t, err)
	require.NotNil(t, td1.ID)

	td2 := &entity.TrustDomain{
		Name: spiffeTD2,
	}
	td2, err = datastore.CreateOrUpdateTrustDomain(ctx, td2)
	require.NoError(t, err)
	require.NotNil(t, td2.ID)

	loc, _ := time.LoadLocation("UTC")
	expiry := time.Now().In(loc).Add(1 * time.Hour)

	// Create first join_token -> trustDomain_1
	req1 := &entity.JoinToken{
		Token:         uuid.NewString(),
		ExpiresAt:     expiry,
		TrustDomainID: td1.ID.UUID,
	}

	token1, err := datastore.CreateJoinToken(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, token1)
	assert.Equal(t, req1.Token, token1.Token)
	assertEqualDate(t, req1.ExpiresAt, token1.ExpiresAt.In(loc))
	require.False(t, token1.Used)
	assert.Equal(t, req1.TrustDomainID, token1.TrustDomainID)

	// Look up token stored in DB and compare
	stored, err := datastore.FindJoinTokensByID(ctx, token1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, token1, stored)

	// Create second join_token -> trustDomain_2
	req2 := &entity.JoinToken{
		Token:         uuid.NewString(),
		ExpiresAt:     expiry,
		TrustDomainID: td2.ID.UUID,
	}

	token2, err := datastore.CreateJoinToken(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, token1)
	assert.Equal(t, req2.Token, token2.Token)
	assert.Equal(t, req1.TrustDomainID, token1.TrustDomainID)
	require.False(t, token2.Used)

	assertEqualDate(t, req2.ExpiresAt, token2.ExpiresAt.In(loc))
	assert.Equal(t, req2.TrustDomainID, token2.TrustDomainID)

	// Look up token stored in DB and compare
	stored, err = datastore.FindJoinTokensByID(ctx, token2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, token2, stored)

	// Create second join_token -> trustDomain_2
	req3 := &entity.JoinToken{
		Token:         uuid.NewString(),
		ExpiresAt:     expiry,
		TrustDomainID: td2.ID.UUID,
	}

	token3, err := datastore.CreateJoinToken(ctx, req3)
	require.NoError(t, err)
	require.NotNil(t, token3)

	// Find tokens by TrustDomainID
	tokens, err := datastore.FindJoinTokensByTrustDomainID(ctx, td1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(tokens))
	require.Contains(t, tokens, token1)

	tokens, err = datastore.FindJoinTokensByTrustDomainID(ctx, td2.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(tokens))
	require.Contains(t, tokens, token2)
	require.Contains(t, tokens, token3)

	// Look up join token by token string
	stored, err = datastore.FindJoinToken(ctx, token1.Token)
	require.NoError(t, err)
	assert.Equal(t, token1, stored)

	stored, err = datastore.FindJoinToken(ctx, token2.Token)
	require.NoError(t, err)
	assert.Equal(t, token2, stored)

	stored, err = datastore.FindJoinToken(ctx, token3.Token)
	require.NoError(t, err)
	assert.Equal(t, token3, stored)

	// List tokens
	tokens, err = datastore.ListJoinTokens(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(tokens))
	require.Contains(t, tokens, token1)
	require.Contains(t, tokens, token2)
	require.Contains(t, tokens, token3)

	// Update join token
	updated, err := datastore.UpdateJoinToken(ctx, token1.ID.UUID, true)
	require.NoError(t, err)
	assert.Equal(t, true, updated.Used)

	// Look up and compare
	stored, err = datastore.FindJoinTokensByID(ctx, token1.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, true, stored.Used)
	assert.Equal(t, updated.UpdatedAt, stored.UpdatedAt)

	// Delete join tokens
	err = datastore.DeleteJoinToken(ctx, token1.ID.UUID)
	require.NoError(t, err)
	stored, err = datastore.FindJoinTokensByID(ctx, token1.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	err = datastore.DeleteJoinToken(ctx, token2.ID.UUID)
	require.NoError(t, err)
	stored, err = datastore.FindJoinTokensByID(ctx, token2.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	err = datastore.DeleteJoinToken(ctx, token3.ID.UUID)
	require.NoError(t, err)
	stored, err = datastore.FindJoinTokensByID(ctx, token3.ID.UUID)
	require.NoError(t, err)
	require.Nil(t, stored)

	tokens, err = datastore.ListJoinTokens(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, len(tokens))
}

func assertEqualDate(t *testing.T, time1 time.Time, time2 time.Time) {
	y1, td1, d1 := time1.Date()
	y2, td2, d2 := time2.Date()
	h1, mt1, s1 := time1.Clock()
	h2, mt2, s2 := time2.Clock()

	require.Equal(t, y1, y2, "Year doesn't match")
	require.Equal(t, td1, td2, "Month doesn't match")
	require.Equal(t, d1, d2, "Day doesn't match")
	require.Equal(t, h1, h2, "Hour doesn't match")
	require.Equal(t, mt1, mt2, "Minute doesn't match")
	require.Equal(t, s1, s2, "Seconds doesn't match")
}

func setupTest(t *testing.T) (*datastore.SQLDatastore, context.Context) {
	ds, err := setupDatastore(t)
	require.NoError(t, err, "Failed to setup the ds")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	return ds, ctx
}

func setupDatastore(t *testing.T) (*datastore.SQLDatastore, error) {
	log := logrus.New()

	conn := startDB(t)
	datastore, err := datastore.NewSQLDatastore(log, conn)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = datastore.Close()
		require.NoError(t, err)
	})
	return datastore, nil
}

// starts a postgres DB in a docker container and returns the connection string
func startDB(tb testing.TB) string {
	pool, err := dockertest.NewPool("")
	require.NoError(tb, err)

	err = pool.Client.Ping()
	require.NoError(tb, err)

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        postgresImage,
		Env: []string{
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_USER=" + user,
			"POSTGRES_DB=" + dbname,
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(tb, err)

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, hostAndPort, dbname)

	tb.Logf("Connecting to a test database on url: %s", databaseUrl)

	// wait until db in container is ready using exponential backoff-retry
	pool.MaxWait = 60 * time.Second
	if err = pool.Retry(func() error {
		db, err := sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		defer func() {
			cerr := db.Close()
			if err != nil {
				err = cerr
			}
		}()

		return db.Ping()
	}); err != nil {
		tb.Fatalf("Could not connect to docker: %s", err)
	}

	tb.Cleanup(func() {
		err = pool.Purge(resource)
		require.NoError(tb, err)
	})

	return databaseUrl
}
