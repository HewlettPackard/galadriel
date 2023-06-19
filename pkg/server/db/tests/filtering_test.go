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

	testCases := []func(*testing.T, context.Context, dbtypes.Engine, func() db.Datastore){
		runListRelationshipsPaginationTest,
		runListRelationshipsFilteringByConsentStatusTest,
		runListRelationshipsFilteringByConsentStatusWithPaginationTest,
		runListRelationshipsOrderByCreatedAtTest,
		runListRelationshipsFilteringByTrustDomainIDTest,
	}

	sqliteDS := func() db.Datastore {
		return setupSQLiteDatastore(t)
	}
	runAllTests(t, ctx, dbtypes.SQLite3, sqliteDS, testCases)

	postgresDS := func() db.Datastore {
		return setupPostgresDatastore(t)
	}
	runAllTests(t, ctx, dbtypes.PostgreSQL, postgresDS, testCases)
}

func TestListTrustDomainsByCriteria(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	testCases := []func(*testing.T, context.Context, dbtypes.Engine, func() db.Datastore){
		runListTrustDomainsPaginationTest,
		runListTrustDomainsOrderByCreatedAtTest,
	}

	sqliteDS := func() db.Datastore {
		return setupSQLiteDatastore(t)
	}
	runAllTests(t, ctx, dbtypes.SQLite3, sqliteDS, testCases)

	postgresDS := func() db.Datastore {
		return setupPostgresDatastore(t)
	}
	runAllTests(t, ctx, dbtypes.PostgreSQL, postgresDS, testCases)
}

func runAllTests(t *testing.T, ctx context.Context, dbType dbtypes.Engine, newDB func() db.Datastore, testCases []func(*testing.T, context.Context, dbtypes.Engine, func() db.Datastore)) {
	for _, testCase := range testCases {
		testCase(t, ctx, dbType, newDB)
	}
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
		relationships, err := ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, relationships, 50)

		listCriteria.PageNumber = 2
		relationships, err = ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, relationships, 50)

		listCriteria.PageNumber = 3
		relationships, err = ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, relationships, 50)

		listCriteria.PageNumber = 4
		relationships, err = ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, relationships, 50)

		listCriteria.PageNumber = 5
		relationships, err = ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, relationships, 0)
	})
}

func runListTrustDomainsPaginationTest(t *testing.T, ctx context.Context, dbType dbtypes.Engine, newDB func() db.Datastore) {
	t.Run(fmt.Sprintf("Test Trust Domain Pagination (%s)", dbType), func(t *testing.T) {
		t.Parallel()
		ds := newDB()
		defer closeDatastore(t, ds)

		createTrustDomains(t, ctx, ds, 200)

		// List trust domains with pagination
		listCriteria := &criteria.ListTrustDomainsCriteria{
			PageNumber: 1,
			PageSize:   50,
		}
		trustDomains, err := ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, trustDomains, 50)

		listCriteria.PageNumber = 2
		trustDomains, err = ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, trustDomains, 50)

		listCriteria.PageNumber = 3
		trustDomains, err = ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, trustDomains, 50)

		listCriteria.PageNumber = 4
		trustDomains, err = ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, trustDomains, 50)

		listCriteria.PageNumber = 5
		trustDomains, err = ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, trustDomains, 0)
	})
}

func runListTrustDomainsOrderByCreatedAtTest(t *testing.T, ctx context.Context, dbType dbtypes.Engine, newDS func() db.Datastore) {
	t.Run(fmt.Sprintf("Test Order by CreatedAt (%s)", dbType), func(t *testing.T) {
		t.Parallel()
		ds := newDS()
		defer closeDatastore(t, ds)

		createTrustDomains(t, ctx, ds, 5)

		// List trust domains ordered by created_at
		listCriteria := &criteria.ListTrustDomainsCriteria{
			OrderByCreatedAt: criteria.OrderAscending,
		}
		trustDomains, err := ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, trustDomains, 5)

		// Convert trustDomains to createdAtProvider instances
		trustDomainAdapters := make([]createdAtProvider, len(trustDomains))
		for i, v := range trustDomains {
			trustDomainAdapters[i] = trustDomainAdapter{td: v}
		}
		assertEntitiesAreInCreatedAtOrder(t, trustDomainAdapters, listCriteria.OrderByCreatedAt)

		// List trust domains ordered by created_at in descending order
		listCriteria.OrderByCreatedAt = criteria.OrderDescending
		trustDomains, err = ds.ListTrustDomains(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, trustDomains, 5)

		for i, v := range trustDomains {
			trustDomainAdapters[i] = trustDomainAdapter{td: v}
		}
		assertEntitiesAreInCreatedAtOrder(t, trustDomainAdapters, listCriteria.OrderByCreatedAt)
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
			relationships, err := ds.ListRelationships(ctx, listCriteria)
			assert.NoError(t, err)
			assert.Equal(t, 200, len(relationships))

			// Assert that the entities have the correct consent status
			assertConsentStatus(t, relationships, filterBy)
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
				relationships, err := ds.ListRelationships(ctx, listCriteria)
				assert.NoError(t, err)

				expectedPageSize := 100
				if i == 3 {
					expectedPageSize = 0
				}
				assert.Equal(t, expectedPageSize, len(relationships))

				assertConsentStatus(t, relationships, filterBy)
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
		relationships, err := ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, relationships, 5)

		relationshipsAdapters := make([]createdAtProvider, len(relationships))
		for i, v := range relationships {
			relationshipsAdapters[i] = relationshipAdapter{rel: v}
		}
		assertEntitiesAreInCreatedAtOrder(t, relationshipsAdapters, criteria.OrderAscending)

		// List relationships ordered by created_at in descending order
		listCriteria.OrderByCreatedAt = criteria.OrderDescending
		relationships, err = ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Len(t, relationships, 5)

		for i, v := range relationships {
			relationshipsAdapters[i] = relationshipAdapter{rel: v}
		}
		assertEntitiesAreInCreatedAtOrder(t, relationshipsAdapters, criteria.OrderDescending)
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
		listRelationships, err := ds.ListRelationships(ctx, listCriteria)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(listRelationships))

		rel := listRelationships[0]
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
		// Create TrustDomains for relationships
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

type createdAtProvider interface {
	CreatedAt() time.Time
}

type trustDomainAdapter struct {
	td *entity.TrustDomain
}

func (a trustDomainAdapter) CreatedAt() time.Time {
	return a.td.CreatedAt
}

type relationshipAdapter struct {
	rel *entity.Relationship
}

func (a relationshipAdapter) CreatedAt() time.Time {
	return a.rel.CreatedAt
}

// assertEntitiesAreInCreatedAtOrder checks if the given entities are in the correct order of creation time.
func assertEntitiesAreInCreatedAtOrder(t *testing.T, entities []createdAtProvider, order criteria.OrderDirection) {
	for i := 0; i < len(entities)-1; i++ {
		createdAt := entities[i].CreatedAt()
		nextCreatedAt := entities[i+1].CreatedAt()

		assertIsInOrder(t, createdAt, nextCreatedAt, order)
	}
}

// assertIsInOrder checks whether the given createdAt times are in the correct order.
func assertIsInOrder(t *testing.T, createdAt, nextCreatedAt time.Time, order criteria.OrderDirection) {
	switch order {
	case criteria.OrderAscending:
		assert.True(t, createdAt.Before(nextCreatedAt) || createdAt.Equal(nextCreatedAt),
			"Expected time %v to be before or at the same time as %v, but it was not.", createdAt, nextCreatedAt)
	case criteria.OrderDescending:
		assert.True(t, createdAt.After(nextCreatedAt) || createdAt.Equal(nextCreatedAt),
			"Expected time %v to be after or at the same time as %v, but it was not.", createdAt, nextCreatedAt)
	case criteria.NoOrder:
		// For NoOrder, we don't perform any check
	default:
		assert.Fail(t, "Unknown order direction: %s", order)
	}
}
