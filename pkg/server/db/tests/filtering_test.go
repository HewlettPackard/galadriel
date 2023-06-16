package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/db"
	"github.com/HewlettPackard/galadriel/pkg/server/db/criteria"
	"github.com/HewlettPackard/galadriel/pkg/server/db/dbtypes"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
)

func TestListRelationshipsByCriteria(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	sqliteDS := func() db.Datastore {
		return setupSQLiteDatastore(t)
	}
	runListRelationshipsPaginationTest(t, ctx, dbtypes.SQLite3, sqliteDS)
	runListRelationshipsFilteringByConsentStatusTest(t, ctx, dbtypes.SQLite3, sqliteDS)
	runListRelationshipsFilteringByConsentStatusWithPaginationTest(t, ctx, dbtypes.SQLite3, sqliteDS)
	runListRelationshipsOrderByCreatedAtTest(t, ctx, dbtypes.SQLite3, sqliteDS)
	runListRelationshipsFilteringByTrustDomainIDTest(t, ctx, dbtypes.SQLite3, sqliteDS)

	postgresDS := func() db.Datastore {
		return setupPostgresDatastore(t)
	}
	runListRelationshipsPaginationTest(t, ctx, dbtypes.PostgreSQL, postgresDS)
	runListRelationshipsFilteringByConsentStatusTest(t, ctx, dbtypes.PostgreSQL, postgresDS)
	runListRelationshipsFilteringByConsentStatusWithPaginationTest(t, ctx, dbtypes.PostgreSQL, sqliteDS)
	runListRelationshipsOrderByCreatedAtTest(t, ctx, dbtypes.PostgreSQL, postgresDS)
	runListRelationshipsFilteringByTrustDomainIDTest(t, ctx, dbtypes.PostgreSQL, postgresDS)
}

func TestListTrustDomainByCriteria(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	sqliteDS := func() db.Datastore {
		return setupSQLiteDatastore(t)
	}
	runListTrustDomainsPaginationTest(t, ctx, dbtypes.SQLite3, sqliteDS)
	runListTrustDomainsOrderByCreatedAtTest(t, ctx, dbtypes.SQLite3, sqliteDS)

	postgresDS := func() db.Datastore {
		return setupPostgresDatastore(t)
	}
	runListTrustDomainsPaginationTest(t, ctx, dbtypes.PostgreSQL, postgresDS)
	runListTrustDomainsOrderByCreatedAtTest(t, ctx, dbtypes.PostgreSQL, postgresDS)
}

func runListRelationshipsPaginationTest(t *testing.T, ctx context.Context, dbType dbtypes.Engine, newDB func() db.Datastore) {
	t.Run(fmt.Sprintf("Test Relationships Pagination (%s)", dbType), func(t *testing.T) {
		t.Parallel()
		ds := newDB()
		defer closeDatastore(t, ds)

		createRelationships(t, ctx, ds, 200)

		// List relationships with pagination
		listCriteria := &criteria.ListRelationshipsCriteria{
			PageNumber: 1,
			PageSize:   50,
		}
		rels, err := ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 50)

		listCriteria.PageNumber = 2
		rels, err = ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 50)

		listCriteria.PageNumber = 3
		rels, err = ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 50)

		listCriteria.PageNumber = 4
		rels, err = ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 50)

		listCriteria.PageNumber = 5
		rels, err = ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 0)
	})
}

func runListTrustDomainsPaginationTest(t *testing.T, ctx context.Context, dbType dbtypes.Engine, newDB func() db.Datastore) {
	t.Run(fmt.Sprintf("Test Trust Domain Pagination (%s)", dbType), func(t *testing.T) {
		t.Parallel()
		ds := newDB()
		defer closeDatastore(t, ds)

		createTrustDomains(t, ctx, ds, 200)

		// List relationships with pagination
		listCriteria := &criteria.ListTrustDomainsCriteria{
			PageNumber: 1,
			PageSize:   50,
		}
		rels, err := ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 50)

		listCriteria.PageNumber = 2
		rels, err = ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 50)

		listCriteria.PageNumber = 3
		rels, err = ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 50)

		listCriteria.PageNumber = 4
		rels, err = ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 50)

		listCriteria.PageNumber = 5
		rels, err = ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 0)
	})
}

func runListTrustDomainsOrderByCreatedAtTest(t *testing.T, ctx context.Context, dbType dbtypes.Engine, newDS func() db.Datastore) {
	t.Run(fmt.Sprintf("Test Order by CreatedAt (%s)", dbType), func(t *testing.T) {
		t.Parallel()
		ds := newDS()
		defer closeDatastore(t, ds)

		createTrustDomains(t, ctx, ds, 5)

		// List relationships ordered by created_at
		listCriteria := &criteria.ListTrustDomainsCriteria{
			OrderByCreatedAt: criteria.OrderAscending,
		}
		rels, err := ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 5)

		assertCreatedAtOrder(t, rels, true)

		// List relationships ordered by created_at in descending order
		listCriteria.OrderByCreatedAt = criteria.OrderDescending
		rels, err = ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 5)

		assertCreatedAtOrder(t, rels, false)
	})
}

func runListRelationshipsFilteringByConsentStatusTest(t *testing.T, ctx context.Context, dbType dbtypes.Engine, newDS func() db.Datastore) {
	t.Run(fmt.Sprintf("Test Filtering By Consent Status (%s)", dbType), func(t *testing.T) {
		t.Parallel()
		ds := newDS()
		defer closeDatastore(t, ds)

		consentStatuses := []entity.ConsentStatus{
			entity.ConsentStatusApproved,
			entity.ConsentStatusDenied,
			entity.ConsentStatusPending,
		}

		createRelationships(t, ctx, ds, 300)

		// List relationships filtered by consent status
		for _, filterBy := range consentStatuses {
			listCriteria := &criteria.ListRelationshipsCriteria{
				FilterByConsentStatus: &filterBy,
			}
			rels, err := ds.ListRelationships(ctx, listCriteria)
			assert.NoError(t, err)
			assert.Equal(t, 200, len(rels))

			// Assert that the entities have the correct consent status
			assertConsentStatus(t, rels, filterBy)
		}
	})
}

func runListRelationshipsFilteringByConsentStatusWithPaginationTest(t *testing.T, ctx context.Context, dbType dbtypes.Engine, newDS func() db.Datastore) {
	t.Run(fmt.Sprintf("Test Filtering and Pagination (%s)", dbType), func(t *testing.T) {
		t.Parallel()
		ds := newDS()
		defer closeDatastore(t, ds)

		consentStatuses := []entity.ConsentStatus{
			entity.ConsentStatusApproved,
			entity.ConsentStatusDenied,
			entity.ConsentStatusPending,
		}

		createRelationships(t, ctx, ds, 300)

		// List relationships filtered by consent status and paginated
		for _, filterBy := range consentStatuses {
			listCriteria := &criteria.ListRelationshipsCriteria{
				FilterByConsentStatus: &filterBy,
				PageSize:              100,
			}

			for i := 1; i <= 3; i++ {
				listCriteria.PageNumber = uint(i)
				rels, err := ds.ListRelationships(ctx, listCriteria)
				assert.NoError(t, err)

				expectedPageSize := 100
				if i == 3 {
					expectedPageSize = 0
				}
				assert.Equal(t, expectedPageSize, len(rels))

				assertConsentStatus(t, rels, filterBy)
			}
		}
	})
}

func runListRelationshipsOrderByCreatedAtTest(t *testing.T, ctx context.Context, dbType dbtypes.Engine, newDS func() db.Datastore) {
	t.Run(fmt.Sprintf("Test Order by CreatedAt (%s)", dbType), func(t *testing.T) {
		t.Parallel()
		ds := newDS()
		defer closeDatastore(t, ds)

		createRelationships(t, ctx, ds, 5)

		// List relationships ordered by created_at
		listCriteria := &criteria.ListRelationshipsCriteria{
			OrderByCreatedAt: criteria.OrderAscending,
		}
		rels, err := ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 5)

		assertCreatedAtOrder(t, rels, true)

		// List relationships ordered by created_at in descending order
		listCriteria.OrderByCreatedAt = criteria.OrderDescending
		rels, err = ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 5)

		assertCreatedAtOrder(t, rels, false)
	})
}

func runListRelationshipsFilteringByTrustDomainIDTest(t *testing.T, ctx context.Context, dbType dbtypes.Engine, newDS func() db.Datastore) {
	t.Run(fmt.Sprintf("Test Filtering By TrustDomain ID (%s)", dbType), func(t *testing.T) {
		t.Parallel()
		ds := newDS()
		defer closeDatastore(t, ds)

		// Create 300 relationships with different TrustDomain IDs
		relationships := createRelationships(t, ctx, ds, 300)

		// Select a TrustDomain ID to filter by
		filterByTrustDomainID := relationships[0].TrustDomainAID

		// List relationships filtered by TrustDomain ID
		listCriteria := &criteria.ListRelationshipsCriteria{
			FilterByTrustDomainID: uuid.NullUUID{Valid: true, UUID: filterByTrustDomainID},
		}
		rels, err := ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(rels))

		rel := rels[0]
		assert.True(t, rel.TrustDomainAID == filterByTrustDomainID || rel.TrustDomainBID == filterByTrustDomainID)
	})
}

func createRelationships(t *testing.T, ctx context.Context, ds db.Datastore, count int) []*entity.Relationship {
	consentStatuses := []entity.ConsentStatus{
		entity.ConsentStatusApproved,
		entity.ConsentStatusDenied,
		entity.ConsentStatusPending,
	}

	relationships := make([]*entity.Relationship, 0, count)
	for i := 0; i < count; i++ {
		// Create TrustDomains
		td1Name := fmt.Sprintf("spiffe://domain%d.com", i*2)
		td1 := &entity.TrustDomain{
			Name: spiffeid.RequireTrustDomainFromString(td1Name),
		}
		td1 = createTrustDomain(ctx, t, ds, td1)

		td2Name := fmt.Sprintf("spiffe://domain%d.com", i*2+1)
		td2 := &entity.TrustDomain{
			Name: spiffeid.RequireTrustDomainFromString(td2Name),
		}
		td2 = createTrustDomain(ctx, t, ds, td2)

		relationship := &entity.Relationship{
			TrustDomainAID: td1.ID.UUID,
			TrustDomainBID: td2.ID.UUID,
		}

		// Set the consent status based on the index
		relationship.TrustDomainAConsent = consentStatuses[i%3]
		relationship.TrustDomainBConsent = consentStatuses[(i+1)%3]

		_, err := ds.CreateOrUpdateRelationship(ctx, relationship)
		assert.NoError(t, err)

		relationships = append(relationships, relationship)
	}

	return relationships
}

func createTrustDomains(t *testing.T, ctx context.Context, ds db.Datastore, count int) []*entity.TrustDomain {

	domains := make([]*entity.TrustDomain, 0, count)
	for i := 0; i < count; i++ {
		// Create TrustDomains
		tdName := fmt.Sprintf("spiffe://domain%d.com", i*2)
		td := &entity.TrustDomain{
			Name:      spiffeid.RequireTrustDomainFromString(tdName),
			CreatedAt: time.Now().Add(time.Duration(i+1) * time.Minute),
		}
		td = createTrustDomain(ctx, t, ds, td)

		domains = append(domains, td)
	}

	return domains
}

func assertConsentStatus(t *testing.T, rels []*entity.Relationship, consentStatus entity.ConsentStatus) {
	for _, rel := range rels {
		assert.True(t, rel.TrustDomainAConsent == consentStatus || rel.TrustDomainBConsent == consentStatus)
	}
}

type TimeComparable interface {
	*entity.Relationship | *entity.TrustDomain
}

func assertCreatedAtOrder[T TimeComparable](t *testing.T, rels []T, ascending bool) {
	for i := 0; i < len(rels)-1; i++ {
		createdAt := timeFromTimeComparable(rels[i])
		nextCreatedAt := timeFromTimeComparable(rels[i+1])

		// +1 means that created is after nextCreatedAt
		// -1 means that created is before nextCreatedAt
		// O is equal, so in the order it doesn't break the order principles
		if ascending {
			assert.NotEqual(t, +1, createdAt.Compare(nextCreatedAt))
		} else {
			assert.NotEqual(t, -1, createdAt.Compare(nextCreatedAt))
		}
	}
}

func timeFromTimeComparable[T TimeComparable](t T) time.Time {
	switch v := any(t).(type) {
	case *entity.TrustDomain:
		return v.CreatedAt
	case *entity.Relationship:
		return v.CreatedAt
	default:
		return time.Time{}
	}
}
