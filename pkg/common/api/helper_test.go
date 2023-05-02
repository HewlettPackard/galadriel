package api

import (
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
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
		harvesterSpiffeID := "spiffe://trust.domain/workload-teste"
		onboardingBundle := "think that I am a bundle"
		trustDomainName := "trust.com"

		td := TrustDomain{
			Id:                uuid.New(),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			Description:       &description,
			Name:              trustDomainName,
			OnboardingBundle:  &onboardingBundle,
			HarvesterSpiffeId: &harvesterSpiffeID,
		}

		etd, err := td.ToEntity()
		assert.NoError(t, err)
		assert.NotNil(t, etd)

		assert.Equal(t, td.Id, etd.ID.UUID)
		assert.Equal(t, td.Name, etd.Name.String())
		assert.Equal(t, td.CreatedAt, etd.CreatedAt)
		assert.Equal(t, td.UpdatedAt, etd.UpdatedAt)
		assert.Equal(t, *td.Description, etd.Description)
		assert.Equal(t, []byte(*td.OnboardingBundle), etd.OnboardingBundle)
		assert.Equal(t, *td.HarvesterSpiffeId, etd.HarvesterSpiffeID.String())
	})
}

func TestTrustDomainFromEntity(t *testing.T) {
	t.Run("Full fill correctly the trust domain API model", func(t *testing.T) {

		uuid := uuid.NullUUID{UUID: uuid.New(), Valid: true}
		description := "a really cool description"
		onboardingBundle := []byte("think that I am a bundle")
		trustDomain := spiffeid.RequireTrustDomainFromString("trust.com")

		harversterSpiffeId, err := spiffeid.FromString("spiffe://trust.domain/workload-teste")
		assert.NoError(t, err)

		etd := entity.TrustDomain{
			ID:                uuid,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			Name:              trustDomain,
			Description:       description,
			OnboardingBundle:  onboardingBundle,
			HarvesterSpiffeID: harversterSpiffeId,
		}

		td := TrustDomainFromEntity(&etd)
		assert.NotNil(t, td)

		assert.Equal(t, etd.ID.UUID, td.Id)
		assert.Equal(t, etd.Name.String(), td.Name)
		assert.Equal(t, etd.CreatedAt, td.CreatedAt)
		assert.Equal(t, etd.UpdatedAt, td.UpdatedAt)
		assert.Equal(t, etd.Description, *td.Description)
		assert.Equal(t, etd.OnboardingBundle, []byte(*td.OnboardingBundle))
		assert.Equal(t, etd.HarvesterSpiffeID.String(), *td.HarvesterSpiffeId)
	})
}

func TestRelationshipFromEntity(t *testing.T) {
	t.Run("Full fill correctly the relationship API model", func(t *testing.T) {

		id := uuid.NullUUID{UUID: uuid.New(), Valid: true}

		eRelationship := entity.Relationship{
			ID:                  id,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
			TrustDomainAID:      uuid.New(),
			TrustDomainBID:      uuid.New(),
			TrustDomainAConsent: true,
			TrustDomainBConsent: false,
		}

		r := RelationshipFromEntity(&eRelationship)
		assert.NotNil(t, r)

		assert.Equal(t, eRelationship.ID.UUID, r.Id)
		assert.Equal(t, eRelationship.CreatedAt, r.CreatedAt)
		assert.Equal(t, eRelationship.UpdatedAt, r.UpdatedAt)
		assert.Equal(t, eRelationship.TrustDomainAID, r.TrustDomainAId)
		assert.Equal(t, eRelationship.TrustDomainBID, r.TrustDomainBId)
		assert.Equal(t, eRelationship.TrustDomainAConsent, r.TrustDomainAConsent)
		assert.Equal(t, eRelationship.TrustDomainBConsent, r.TrustDomainBConsent)
	})
}
