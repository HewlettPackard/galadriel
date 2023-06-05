package db

import (
	"context"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/test/fakes/fakedatastore"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
)

func TestPopulateTrustDomainNames(t *testing.T) {
	ctx := context.Background()
	db := fakedatastore.NewFakeDB()

	tdA := &entity.TrustDomain{Name: spiffeid.RequireTrustDomainFromString("td-a.org"), ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	tdB := &entity.TrustDomain{Name: spiffeid.RequireTrustDomainFromString("td-b.org"), ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	tdC := &entity.TrustDomain{Name: spiffeid.RequireTrustDomainFromString("td-c.org"), ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	rel1 := &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdB.ID.UUID, TrustDomainAConsent: entity.ConsentStatusPending, TrustDomainBConsent: entity.ConsentStatusPending, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	rel2 := &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdC.ID.UUID, TrustDomainAConsent: entity.ConsentStatusPending, TrustDomainBConsent: entity.ConsentStatusPending, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	rel3 := &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdB.ID.UUID, TrustDomainAConsent: entity.ConsentStatusApproved, TrustDomainBConsent: entity.ConsentStatusPending, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	rels := []*entity.Relationship{rel1, rel2, rel3}
	db.WithTrustDomains(tdA, tdB, tdC)
	db.WithRelationships(rels...)

	updatedRelationships, err := PopulateTrustDomainNames(ctx, db, rels...)
	assert.NoError(t, err)

	for _, r := range updatedRelationships {
		tda, _ := db.FindTrustDomainByID(ctx, r.TrustDomainAID)
		assert.Equal(t, tda.Name, r.TrustDomainAName)

		tdb, _ := db.FindTrustDomainByID(ctx, r.TrustDomainBID)
		assert.Equal(t, tdb.Name, r.TrustDomainBName)
	}
}
