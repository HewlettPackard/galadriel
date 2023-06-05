package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/db"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	spiffeTD1 = spiffeid.RequireTrustDomainFromString("foo.test")
	spiffeTD2 = spiffeid.RequireTrustDomainFromString("bar.test")
	spiffeTD3 = spiffeid.RequireTrustDomainFromString("baz.test")
)

func TestSuite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	sqliteDS := func() db.Datastore {
		return setupSQLiteDatastore(t)
	}
	runTests(t, ctx, sqliteDS)

	postgresDS := func() db.Datastore {
		return setupPostgresDatastore(t)
	}
	runTests(t, ctx, postgresDS)
}

func runTests(t *testing.T, ctx context.Context, newDS func() db.Datastore) {
	t.Run("Test CRUD TrustDomains", func(t *testing.T) {
		t.Parallel()
		ds := newDS()
		defer closeDatastore(t, ds)

		// Create trust domain
		req1 := &entity.TrustDomain{
			Name: spiffeTD1,
		}
		td1, err := ds.CreateOrUpdateTrustDomain(ctx, req1)
		assert.NoError(t, err)
		assert.NotNil(t, td1.ID)
		assert.Equal(t, req1.Name, td1.Name)
		assert.NotNil(t, td1.CreatedAt)
		assert.NotNil(t, td1.UpdatedAt)

		// Create second trust domain
		req2 := &entity.TrustDomain{
			Name: spiffeTD2,
		}
		td2, err := ds.CreateOrUpdateTrustDomain(ctx, req2)
		assert.NoError(t, err)
		assert.NotNil(t, td2.ID)
		assert.Equal(t, req2.Name, td2.Name)
		assert.NotNil(t, td2.CreatedAt)
		assert.NotNil(t, td2.UpdatedAt)

		// Find trust domain by ID
		stored, err := ds.FindTrustDomainByID(ctx, td1.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, td1, stored)
		stored, err = ds.FindTrustDomainByID(ctx, td2.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, td2, stored)

		// Update trust domain
		td1.Description = "updated_description"

		updated1, err := ds.CreateOrUpdateTrustDomain(ctx, td1)
		assert.NoError(t, err)
		assert.NotNil(t, updated1)

		// Look up trust domain stored in DB and compare
		stored, err = ds.FindTrustDomainByID(ctx, td1.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, td1.ID, stored.ID)
		assert.Equal(t, td1.Description, stored.Description)

		// Find trust domain by name
		td1 = updated1
		stored, err = ds.FindTrustDomainByName(ctx, td1.Name)
		assert.NoError(t, err)
		assert.Equal(t, td1, stored)

		// List all trust domains
		list, err := ds.ListTrustDomains(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(list))
		assert.Contains(t, list, td1)
		assert.Contains(t, list, td2)

		// Delete trust domain
		err = ds.DeleteTrustDomain(ctx, td1.ID.UUID)
		assert.NoError(t, err)
		stored, err = ds.FindTrustDomainByID(ctx, td1.ID.UUID)
		assert.NoError(t, err)
		require.Nil(t, stored)
	})
	t.Run("Test TrustDomain Unique Constraint", func(t *testing.T) {
		t.Parallel()
		ds := newDS()
		defer closeDatastore(t, ds)

		td1 := &entity.TrustDomain{
			Name: spiffeTD1,
		}
		_, err := ds.CreateOrUpdateTrustDomain(ctx, td1)
		assert.NoError(t, err)

		// second trustDomain with same trust domain
		td2 := &entity.TrustDomain{
			Name: spiffeTD1,
		}
		_, err = ds.CreateOrUpdateTrustDomain(ctx, td2)
		require.Error(t, err)

		sqliteExpectedErr := "UNIQUE constraint failed"
		postgresExpectedErr := "duplicate key value violates unique constraint"
		assertErrorString(t, err, sqliteExpectedErr, postgresExpectedErr)
	})
	t.Run("Test CRUD Relationships", func(t *testing.T) {
		t.Parallel()
		ds := newDS()
		defer closeDatastore(t, ds)

		// Create TrustDomains
		td1 := &entity.TrustDomain{
			Name: spiffeTD1,
		}
		td1 = createTrustDomain(ctx, t, ds, td1)

		td2 := &entity.TrustDomain{
			Name: spiffeTD2,
		}
		td2 = createTrustDomain(ctx, t, ds, td2)

		td3 := &entity.TrustDomain{
			Name: spiffeTD3,
		}
		td3 = createTrustDomain(ctx, t, ds, td3)

		// Create relationship TrustDomain1 -- TrustDomain2
		req1 := &entity.Relationship{
			TrustDomainAID: td1.ID.UUID,
			TrustDomainBID: td2.ID.UUID,
		}

		relationship1, err := ds.CreateOrUpdateRelationship(ctx, req1)
		assert.NoError(t, err)
		assert.NotNil(t, relationship1.ID)
		assert.NotNil(t, relationship1.CreatedAt)
		assert.NotNil(t, relationship1.UpdatedAt)
		assert.Equal(t, req1.TrustDomainAID, relationship1.TrustDomainAID)
		assert.Equal(t, req1.TrustDomainBID, relationship1.TrustDomainBID)
		assert.Equal(t, entity.ConsentStatusPending, relationship1.TrustDomainAConsent)
		assert.Equal(t, entity.ConsentStatusPending, relationship1.TrustDomainBConsent)

		// Create relationship TrustDomain2 -- TrustDomain3
		req2 := &entity.Relationship{
			TrustDomainAID: td2.ID.UUID,
			TrustDomainBID: td3.ID.UUID,
		}

		relationship2, err := ds.CreateOrUpdateRelationship(ctx, req2)
		assert.NoError(t, err)
		assert.NotNil(t, relationship2.ID)
		assert.NotNil(t, relationship2.CreatedAt)
		assert.NotNil(t, relationship2.UpdatedAt)
		assert.Equal(t, req2.TrustDomainAID, relationship2.TrustDomainAID)
		assert.Equal(t, req2.TrustDomainBID, relationship2.TrustDomainBID)
		assert.Equal(t, entity.ConsentStatusPending, relationship2.TrustDomainAConsent)
		assert.Equal(t, entity.ConsentStatusPending, relationship2.TrustDomainBConsent)

		// Find relationship by ID
		stored, err := ds.FindRelationshipByID(ctx, relationship1.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, relationship1, stored)
		stored, err = ds.FindRelationshipByID(ctx, relationship2.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, relationship2, stored)

		// Update relationship
		relationship1.TrustDomainAConsent = entity.ConsentStatusApproved
		relationship1.TrustDomainBConsent = entity.ConsentStatusDenied
		updated1, err := ds.CreateOrUpdateRelationship(ctx, relationship1)
		assert.NoError(t, err)
		assert.Equal(t, relationship1.TrustDomainAConsent, updated1.TrustDomainAConsent)
		assert.Equal(t, relationship1.TrustDomainBConsent, updated1.TrustDomainBConsent)
		relationship1 = updated1

		// Find relationship by trust domain IDs
		rels, err := ds.FindRelationshipsByTrustDomainID(ctx, td2.ID.UUID)
		assert.NoError(t, err)
		assert.Len(t, rels, 2)
		assert.Contains(t, rels, relationship1)
		assert.Contains(t, rels, relationship2)

		rels, err = ds.FindRelationshipsByTrustDomainID(ctx, td1.ID.UUID)
		assert.NoError(t, err)
		assert.Len(t, rels, 1)
		assert.Contains(t, rels, relationship1)

		// List all relationships
		rels, err = ds.ListRelationships(ctx)
		assert.NoError(t, err)
		assert.Len(t, rels, 2)

		// Delete relationship
		err = ds.DeleteRelationship(ctx, relationship1.ID.UUID)
		assert.NoError(t, err)
		stored, err = ds.FindRelationshipByID(ctx, relationship1.ID.UUID)
		assert.NoError(t, err)
		assert.Nil(t, stored)

		err = ds.DeleteRelationship(ctx, relationship2.ID.UUID)
		assert.NoError(t, err)
		stored, err = ds.FindRelationshipByID(ctx, relationship2.ID.UUID)
		assert.NoError(t, err)
		assert.Nil(t, stored)
	})
	t.Run("Test Relationship ForeignKey Constraints", func(t *testing.T) {
		t.Parallel()
		ds := newDS()
		defer closeDatastore(t, ds)

		td1 := &entity.TrustDomain{
			Name: spiffeTD1,
		}
		td1, err := ds.CreateOrUpdateTrustDomain(ctx, td1)
		assert.NoError(t, err)

		td2 := &entity.TrustDomain{
			Name: spiffeTD2,
		}
		td2, err = ds.CreateOrUpdateTrustDomain(ctx, td2)
		assert.NoError(t, err)

		relationship1 := &entity.Relationship{
			TrustDomainAID: td1.ID.UUID,
			TrustDomainBID: td2.ID.UUID,
		}
		relationship1, err = ds.CreateOrUpdateRelationship(ctx, relationship1)
		assert.NoError(t, err)

		// Cannot add a new relationship for the same TrustDomains
		relationship1.ID = uuid.NullUUID{}
		_, err = ds.CreateOrUpdateRelationship(ctx, relationship1)
		require.Error(t, err)

		sqliteExpectedError := "UNIQUE constraint failed"
		postgresExpectedError := "duplicate key value violates unique constraint"
		assertErrorString(t, err, sqliteExpectedError, postgresExpectedError)

		// Cannot delete Trust Domain that has a relationship associated
		err = ds.DeleteTrustDomain(ctx, td1.ID.UUID)
		require.Error(t, err)

		sqliteExpectedError = "FOREIGN KEY constraint failed"
		postgresExpectedError = "violates foreign key constraint"
		assertErrorString(t, err, sqliteExpectedError, postgresExpectedError)

		// Cannot delete Trust Domain that has a relationship associated
		err = ds.DeleteTrustDomain(ctx, td2.ID.UUID)
		require.Error(t, err)
		assertErrorString(t, err, sqliteExpectedError, postgresExpectedError)
	})
	t.Run("Test CRUD Bundles", func(t *testing.T) {
		t.Parallel()
		ds := newDS()
		defer closeDatastore(t, ds)

		// Create trustDomains to associate the bundles
		td1 := &entity.TrustDomain{
			Name: spiffeTD1,
		}
		td1, err := ds.CreateOrUpdateTrustDomain(ctx, td1)
		assert.NoError(t, err)
		assert.NotNil(t, td1.ID)

		td2 := &entity.TrustDomain{
			Name: spiffeTD2,
		}
		td2, err = ds.CreateOrUpdateTrustDomain(ctx, td2)
		assert.NoError(t, err)
		assert.NotNil(t, td2.ID)

		// Create first Data - trustDomain-1
		req1 := &entity.Bundle{
			Data:               []byte{1, 2, 3},
			Digest:             []byte("test-digest"),
			Signature:          []byte{4, 2},
			SigningCertificate: []byte{50, 60},
			TrustDomainID:      td1.ID.UUID,
		}

		b1, err := ds.CreateOrUpdateBundle(ctx, req1)
		assert.NoError(t, err)
		assert.NotNil(t, b1)
		assert.Equal(t, req1.Data, b1.Data)
		assert.Equal(t, req1.Digest, b1.Digest)
		assert.Equal(t, req1.Signature, b1.Signature)
		assert.Equal(t, req1.SigningCertificate, b1.SigningCertificate)
		assert.Equal(t, req1.TrustDomainID, b1.TrustDomainID)

		// Look up bundle stored in DB and compare
		stored, err := ds.FindBundleByID(ctx, b1.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, b1, stored)

		// Create second Data -> trustDomain-2
		req2 := &entity.Bundle{
			Data:               []byte{10, 20, 30},
			Digest:             []byte("test-digest-2"),
			Signature:          []byte{40, 20},
			SigningCertificate: []byte{80, 90},
			TrustDomainID:      td2.ID.UUID,
		}

		b2, err := ds.CreateOrUpdateBundle(ctx, req2)
		assert.NoError(t, err)
		assert.NotNil(t, b1)
		assert.Equal(t, req2.Data, b2.Data)
		assert.Equal(t, req2.Digest, b2.Digest)
		assert.Equal(t, req2.Signature, b2.Signature)
		assert.Equal(t, req2.SigningCertificate, b2.SigningCertificate)
		assert.Equal(t, req2.TrustDomainID, b2.TrustDomainID)

		// Look up bundle stored in DB and compare
		stored, err = ds.FindBundleByID(ctx, b2.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, b2, stored)

		// Find bundles by TrustDomainID
		stored, err = ds.FindBundleByTrustDomainID(ctx, td1.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, b1, stored)

		stored, err = ds.FindBundleByTrustDomainID(ctx, td2.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, b2, stored)

		// Update Data
		b1.Data = []byte{'a', 'b', 'c'}
		b1.Digest = []byte("test-digest-3")
		b1.Signature = []byte{'f', 'g', 'h'}
		b1.SigningCertificate = []byte{'f', 'g', 'h'}

		updated, err := ds.CreateOrUpdateBundle(ctx, b1)
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, b1.Data, updated.Data)
		assert.Equal(t, b1.Digest, updated.Digest)
		assert.Equal(t, b1.Signature, updated.Signature)
		assert.Equal(t, b1.SigningCertificate, updated.SigningCertificate)
		assert.Equal(t, b1.TrustDomainID, updated.TrustDomainID)

		// Look up bundle stored in DB and compare
		stored, err = ds.FindBundleByID(ctx, b1.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, updated, stored)

		// List bundles
		bundles, err := ds.ListBundles(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(bundles))
		require.Contains(t, bundles, updated)
		require.Contains(t, bundles, b2)

		// Delete first bundle
		err = ds.DeleteBundle(ctx, b1.ID.UUID)
		assert.NoError(t, err)

		// Look up deleted bundle
		stored, err = ds.FindBundleByID(ctx, b1.ID.UUID)
		assert.NoError(t, err)
		require.Nil(t, stored)

		// Delete second bundle
		err = ds.DeleteBundle(ctx, b2.ID.UUID)
		assert.NoError(t, err)

		// Look up deleted bundle
		stored, err = ds.FindBundleByID(ctx, b2.ID.UUID)
		assert.NoError(t, err)
		require.Nil(t, stored)
	})
	t.Run("Test Bundle Unique TrustDomain Constraint", func(t *testing.T) {
		t.Parallel()
		ds := newDS()
		defer closeDatastore(t, ds)

		// Create trustDomain to associate the bundles
		td1 := &entity.TrustDomain{
			Name: spiffeTD1,
		}
		td1, err := ds.CreateOrUpdateTrustDomain(ctx, td1)
		assert.NoError(t, err)
		assert.NotNil(t, td1.ID)

		// Create Data
		b1 := &entity.Bundle{
			Data:               []byte{1, 2, 3},
			Digest:             []byte("test-digest-1"),
			Signature:          []byte{4, 2},
			SigningCertificate: []byte{50, 60},
			TrustDomainID:      td1.ID.UUID,
		}
		b1, err = ds.CreateOrUpdateBundle(ctx, b1)
		assert.NoError(t, err)
		assert.NotNil(t, b1)

		// Create second Data associated to same trustDomain
		b2 := &entity.Bundle{
			Data:               []byte{10, 20, 30},
			Digest:             []byte("test-digest-2"),
			Signature:          []byte{40, 20},
			SigningCertificate: []byte{80, 90},
			TrustDomainID:      td1.ID.UUID,
		}
		b2, err = ds.CreateOrUpdateBundle(ctx, b2)
		require.Error(t, err)
		require.Nil(t, b2)

		sqliteExpectedErr := "UNIQUE constraint failed"
		postgresExpectedErr := "duplicate key value violates unique constraint"
		assertErrorString(t, err, sqliteExpectedErr, postgresExpectedErr)
	})
	t.Run("Test CRUD Join Tokens", func(t *testing.T) {
		t.Parallel()
		ds := newDS()
		defer closeDatastore(t, ds)

		// Create trustDomains to associate the join tokens
		td1 := &entity.TrustDomain{
			Name: spiffeTD1,
		}
		td1, err := ds.CreateOrUpdateTrustDomain(ctx, td1)
		assert.NoError(t, err)
		assert.NotNil(t, td1.ID)

		td2 := &entity.TrustDomain{
			Name: spiffeTD2,
		}
		td2, err = ds.CreateOrUpdateTrustDomain(ctx, td2)
		assert.NoError(t, err)
		assert.NotNil(t, td2.ID)

		loc, _ := time.LoadLocation("UTC")
		expiry := time.Now().In(loc).Add(1 * time.Hour)

		// Create first join_token -> trustDomain_1
		req1 := &entity.JoinToken{
			Token:         uuid.NewString(),
			ExpiresAt:     expiry,
			TrustDomainID: td1.ID.UUID,
		}

		token1, err := ds.CreateJoinToken(ctx, req1)
		assert.NoError(t, err)
		assert.NotNil(t, token1)
		assert.Equal(t, req1.Token, token1.Token)
		assertEqualDate(t, req1.ExpiresAt, token1.ExpiresAt.In(loc))
		require.False(t, token1.Used)
		assert.Equal(t, req1.TrustDomainID, token1.TrustDomainID)

		// Look up token stored in DB and compare
		stored, err := ds.FindJoinTokensByID(ctx, token1.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, token1, stored)

		// Create second join_token -> trustDomain_2
		req2 := &entity.JoinToken{
			Token:         uuid.NewString(),
			ExpiresAt:     expiry,
			TrustDomainID: td2.ID.UUID,
		}

		token2, err := ds.CreateJoinToken(ctx, req2)
		assert.NoError(t, err)
		assert.NotNil(t, token1)
		assert.Equal(t, req2.Token, token2.Token)
		assert.Equal(t, req1.TrustDomainID, token1.TrustDomainID)
		require.False(t, token2.Used)

		assertEqualDate(t, req2.ExpiresAt, token2.ExpiresAt.In(loc))
		assert.Equal(t, req2.TrustDomainID, token2.TrustDomainID)

		// Look up token stored in DB and compare
		stored, err = ds.FindJoinTokensByID(ctx, token2.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, token2, stored)

		// Create second join_token -> trustDomain_2
		req3 := &entity.JoinToken{
			Token:         uuid.NewString(),
			ExpiresAt:     expiry,
			TrustDomainID: td2.ID.UUID,
		}

		token3, err := ds.CreateJoinToken(ctx, req3)
		assert.NoError(t, err)
		assert.NotNil(t, token3)

		// Find tokens by TrustDomainID
		tokens, err := ds.FindJoinTokensByTrustDomainID(ctx, td1.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(tokens))
		require.Contains(t, tokens, token1)

		tokens, err = ds.FindJoinTokensByTrustDomainID(ctx, td2.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(tokens))
		require.Contains(t, tokens, token2)
		require.Contains(t, tokens, token3)

		// Look up join token by token string
		stored, err = ds.FindJoinToken(ctx, token1.Token)
		assert.NoError(t, err)
		assert.Equal(t, token1, stored)

		stored, err = ds.FindJoinToken(ctx, token2.Token)
		assert.NoError(t, err)
		assert.Equal(t, token2, stored)

		stored, err = ds.FindJoinToken(ctx, token3.Token)
		assert.NoError(t, err)
		assert.Equal(t, token3, stored)

		// List tokens
		tokens, err = ds.ListJoinTokens(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(tokens))
		require.Contains(t, tokens, token1)
		require.Contains(t, tokens, token2)
		require.Contains(t, tokens, token3)

		// Update join token
		updated, err := ds.UpdateJoinToken(ctx, token1.ID.UUID, true)
		assert.NoError(t, err)
		assert.Equal(t, true, updated.Used)

		// Look up and compare
		stored, err = ds.FindJoinTokensByID(ctx, token1.ID.UUID)
		assert.NoError(t, err)
		assert.Equal(t, true, stored.Used)
		assert.Equal(t, updated.UpdatedAt, stored.UpdatedAt)

		// Delete join tokens
		err = ds.DeleteJoinToken(ctx, token1.ID.UUID)
		assert.NoError(t, err)
		stored, err = ds.FindJoinTokensByID(ctx, token1.ID.UUID)
		assert.NoError(t, err)
		require.Nil(t, stored)

		err = ds.DeleteJoinToken(ctx, token2.ID.UUID)
		assert.NoError(t, err)
		stored, err = ds.FindJoinTokensByID(ctx, token2.ID.UUID)
		assert.NoError(t, err)
		require.Nil(t, stored)

		err = ds.DeleteJoinToken(ctx, token3.ID.UUID)
		assert.NoError(t, err)
		stored, err = ds.FindJoinTokensByID(ctx, token3.ID.UUID)
		assert.NoError(t, err)
		require.Nil(t, stored)

		tokens, err = ds.ListJoinTokens(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(tokens))
	})
}

func createTrustDomain(ctx context.Context, t *testing.T, ds db.Datastore, req *entity.TrustDomain) *entity.TrustDomain {
	td1, err := ds.CreateOrUpdateTrustDomain(ctx, req)
	require.NoError(t, err)
	return td1
}

func closeDatastore(t *testing.T, ds db.Datastore) {
	switch d := ds.(type) {
	case interface {
		Close() error
	}:
		if err := d.Close(); err != nil {
			t.Errorf("error closing datastore: %v", err)
		}
	}
}

// assertErrorString asserts that the error string is one of the expected error strings
func assertErrorString(t *testing.T, err error, s1, s2 string) {
	if err == nil {
		t.Fatalf("expected error containing either '%s' or '%s', but got no error", s1, s2)
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, s1) && !strings.Contains(errMsg, s2) {
		t.Fatalf("expected error containing either '%s' or '%s', but got '%s'", s1, s2, errMsg)
	}
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
