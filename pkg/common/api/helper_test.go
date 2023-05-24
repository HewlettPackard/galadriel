package api

import (
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrustDomainToEntity(t *testing.T) {
	t.Run("Does not allow wrong spiffe ids", func(t *testing.T) {
		harvesterSpiffeID := "Not and spiffeid"
		td := TrustDomain{HarvesterSpiffeId: &harvesterSpiffeID}
		etd, err := td.ToEntity()
		assert.Error(t, err)
		assert.Nil(t, etd)
	})

	t.Run("Does not allow wrong trust domain", func(t *testing.T) {
		harvesterSpiffeID := "spiffe://trust.domain/workload-teste"
		td := TrustDomain{
			HarvesterSpiffeId: &harvesterSpiffeID,
			Name:              "Not a Trust Domain",
		}

		etd, err := td.ToEntity()
		assert.Error(t, err)
		assert.Nil(t, etd)
	})

	t.Run("Full fill correctly the entity model", func(t *testing.T) {
		description := "A description"
		trustDomainName := "trust.com"

		td := TrustDomain{
			Id:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Description: &description,
			Name:        trustDomainName,
		}

		etd, err := td.ToEntity()
		assert.NoError(t, err)
		assert.NotNil(t, etd)

		assert.Equal(t, td.Id, etd.ID.UUID)
		assert.Equal(t, td.Name, etd.Name.String())
		assert.Equal(t, td.CreatedAt, etd.CreatedAt)
		assert.Equal(t, td.UpdatedAt, etd.UpdatedAt)
		assert.Equal(t, *td.Description, etd.Description)
	})
}

func TestTrustDomainFromEntity(t *testing.T) {
	uuid := uuid.NullUUID{UUID: uuid.New(), Valid: true}
	description := "a really cool description"
	trustDomain := spiffeid.RequireTrustDomainFromString("trust.com")

	etd := entity.TrustDomain{
		ID:          uuid,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Name:        trustDomain,
		Description: description,
	}

	td := TrustDomainFromEntity(&etd)
	assert.NotNil(t, td)

	assert.Equal(t, etd.ID.UUID, td.Id)
	assert.Equal(t, etd.Name.String(), td.Name)
	assert.Equal(t, etd.CreatedAt, td.CreatedAt)
	assert.Equal(t, etd.UpdatedAt, td.UpdatedAt)
	assert.Equal(t, etd.Description, *td.Description)
}

func TestRelationshipToEntity(t *testing.T) {
	// Arrange
	id := uuid.New()
	trustDomainAName := "td1"
	trustDomainBName := "td2"
	trustDomainAId := uuid.New()
	trustDomainBId := uuid.New()

	r := Relationship{
		Id:                  id,
		TrustDomainAName:    &trustDomainAName,
		TrustDomainBName:    &trustDomainBName,
		TrustDomainAId:      trustDomainAId,
		TrustDomainBId:      trustDomainBId,
		TrustDomainAConsent: Approved,
		TrustDomainBConsent: Denied,
	}

	// Act
	ent, err := r.ToEntity()

	// Assert
	require.NoError(t, err)
	require.Equal(t, id, ent.ID.UUID)
	require.Equal(t, trustDomainAId, ent.TrustDomainAID)
	require.Equal(t, trustDomainBId, ent.TrustDomainBID)
	require.Equal(t, trustDomainAName, ent.TrustDomainAName.String())
	require.Equal(t, trustDomainBName, ent.TrustDomainBName.String())
	require.Equal(t, entity.ConsentStatusApproved, ent.TrustDomainAConsent)
	require.Equal(t, entity.ConsentStatusDenied, ent.TrustDomainBConsent)

	// Test invalid trust domain A name
	invalidTrustDomainAName := "invalid trust domain"
	r.TrustDomainAName = &invalidTrustDomainAName
	_, err = r.ToEntity()
	require.Error(t, err)

	// Test invalid trust domain B name
	r.TrustDomainAName = &trustDomainAName
	invalidTrustDomainBName := "invalid trust domain"
	r.TrustDomainBName = &invalidTrustDomainBName
	_, err = r.ToEntity()
	require.Error(t, err)
}

func TestRelationshipFromEntity(t *testing.T) {
	id := uuid.NullUUID{UUID: uuid.New(), Valid: true}

	eRelationship := entity.Relationship{
		ID:                  id,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		TrustDomainAID:      uuid.New(),
		TrustDomainBID:      uuid.New(),
		TrustDomainAConsent: entity.ConsentStatusPending,
		TrustDomainBConsent: entity.ConsentStatusApproved,
	}

	r := RelationshipFromEntity(&eRelationship)
	assert.NotNil(t, r)

	assert.Equal(t, eRelationship.ID.UUID, r.Id)
	assert.Equal(t, eRelationship.CreatedAt, r.CreatedAt)
	assert.Equal(t, eRelationship.UpdatedAt, r.UpdatedAt)
	assert.Equal(t, eRelationship.TrustDomainAID, r.TrustDomainAId)
	assert.Equal(t, eRelationship.TrustDomainBID, r.TrustDomainBId)
	assert.Equal(t, string(eRelationship.TrustDomainAConsent), string(r.TrustDomainAConsent))
	assert.Equal(t, string(eRelationship.TrustDomainBConsent), string(r.TrustDomainBConsent))
}

func TestMapRelationships(t *testing.T) {
	relationships := []*entity.Relationship{
		{
			ID:                  uuid.NullUUID{UUID: uuid.New(), Valid: true},
			TrustDomainAID:      uuid.New(),
			TrustDomainBID:      uuid.New(),
			TrustDomainAConsent: "approved",
			TrustDomainBConsent: "approved",
		},
		{
			ID:                  uuid.NullUUID{UUID: uuid.New(), Valid: true},
			TrustDomainAID:      uuid.New(),
			TrustDomainBID:      uuid.New(),
			TrustDomainAConsent: "denied",
			TrustDomainBConsent: "approved",
		},
	}

	// Call MapRelationships
	cRelationships := MapRelationships(relationships...)

	// Verify results
	assert.Equal(t, len(relationships), len(cRelationships))
	for i, cRelation := range cRelationships {
		assert.Equal(t, relationships[i].ID.UUID, cRelation.Id)
		assert.Equal(t, relationships[i].TrustDomainAID, cRelation.TrustDomainAId)
		assert.Equal(t, relationships[i].TrustDomainBID, cRelation.TrustDomainBId)
		assert.Equal(t, ConsentStatus(relationships[i].TrustDomainAConsent), cRelation.TrustDomainAConsent)
		assert.Equal(t, ConsentStatus(relationships[i].TrustDomainBConsent), cRelation.TrustDomainBConsent)
	}
}
