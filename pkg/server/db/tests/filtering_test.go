package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/db"
	"github.com/HewlettPackard/galadriel/pkg/server/db/options"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
)

func TestFilteringSuite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	sqliteDS := func() db.Datastore {
		return setupSQLiteDatastore(t)
	}
	runPaginationTest(t, ctx, db.SQLite, sqliteDS)
	runFilteringByConsentStatusTest(t, ctx, db.SQLite, sqliteDS)

	postgresDS := func() db.Datastore {
		return setupPostgresDatastore(t)
	}
	runPaginationTest(t, ctx, db.Postgres, postgresDS)
	runFilteringByConsentStatusTest(t, ctx, db.Postgres, postgresDS)
}

func runPaginationTest(t *testing.T, ctx context.Context, dbType db.Engine, newDB func() db.Datastore) {
	t.Run(fmt.Sprintf("Test Relationships Pagination (%s)", dbType), func(t *testing.T) {
		t.Parallel()
		ds := newDB()
		defer closeDatastore(t, ds)

		// Create 200 relationships
		for i := 0; i < 200; i++ {
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

			req := &entity.Relationship{
				TrustDomainAID: td1.ID.UUID,
				TrustDomainBID: td2.ID.UUID,
			}
			_, err := ds.CreateOrUpdateRelationship(ctx, req)
			assert.NoError(t, err)
		}

		// List relationships with pagination
		criteria := &options.ListRelationshipsCriteria{
			PageNumber: 1,
			PageSize:   50,
		}
		rels, err := ds.ListRelationships(ctx, criteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 50)

		criteria.PageNumber = 2
		rels, err = ds.ListRelationships(ctx, criteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 50)

		criteria.PageNumber = 3
		rels, err = ds.ListRelationships(ctx, criteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 50)

		criteria.PageNumber = 4
		rels, err = ds.ListRelationships(ctx, criteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 50)

		criteria.PageNumber = 5
		rels, err = ds.ListRelationships(ctx, criteria)
		assert.NoError(t, err)
		assert.Len(t, rels, 0)
	})
}

func runFilteringByConsentStatusTest(t *testing.T, ctx context.Context, dbType db.Engine, newDB func() db.Datastore) {
	t.Run(fmt.Sprintf("Test Filtering By Consent Status (%s)", dbType), func(t *testing.T) {
		t.Parallel()
		ds := newDB()
		defer closeDatastore(t, ds)

		consentStatuses := []entity.ConsentStatus{
			entity.ConsentStatusApproved,
			entity.ConsentStatusDenied,
			entity.ConsentStatusPending,
		}

		// Create 300 relationships with different consent statuses
		for i := 0; i < 300; i++ {
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

			req := &entity.Relationship{
				TrustDomainAID: td1.ID.UUID,
				TrustDomainBID: td2.ID.UUID,
			}

			rel, err := ds.CreateOrUpdateRelationship(ctx, req)
			assert.NoError(t, err)

			rel.TrustDomainAConsent = consentStatuses[i%3]
			rel.TrustDomainBConsent = consentStatuses[(i+1)%3]
			_, err = ds.CreateOrUpdateRelationship(ctx, rel)
			assert.NoError(t, err)
		}

		// List relationships filtered by consent status
		for _, filterBy := range consentStatuses {
			criteria := &options.ListRelationshipsCriteria{
				FilterByConsentStatus: &filterBy,
			}
			rels, err := ds.ListRelationships(ctx, criteria)
			assert.NoError(t, err)
			// Given the pattern of consent statuses, we expect to get back two-thirds of the total number of relationships
			assert.Equal(t, 200, len(rels))
		}
	})
}
