package admin

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRelationshipRequestToEntity(t *testing.T) {
	t.Run("Full fill correctly the relationship entity model", func(t *testing.T) {
		releationshipRequest := RelationshipRequest{
			TrustDomainAId: uuid.New(),
			TrustDomainBId: uuid.New(),
		}

		r := releationshipRequest.ToEntity()
		assert.NotNil(t, r)

		assert.Equal(t, releationshipRequest.TrustDomainAId, r.TrustDomainAID)
		assert.Equal(t, releationshipRequest.TrustDomainBId, r.TrustDomainBID)
	})
}

func TestTrustDomainPutToEntity(t *testing.T) {
	t.Run("Does not allow wrong trust domain names", func(t *testing.T) {
		description := "a cool description"

		tdPut := TrustDomainPut{
			Name:        "A wrong trust domain name",
			Description: &description,
		}

		trustDomain, err := tdPut.ToEntity()
		assert.ErrorContains(t, err, "malformed trust domain[A wrong trust domain name]")
		assert.Nil(t, trustDomain)
	})

	t.Run("Full fill correctly the trust domain entity model", func(t *testing.T) {
		description := "a cool description"

		tdPut := TrustDomainPut{
			Name:        "trust.com",
			Description: &description,
		}

		trustDomain, err := tdPut.ToEntity()
		assert.NoError(t, err)
		assert.NotNil(t, trustDomain)

		assert.Equal(t, tdPut.Name, trustDomain.Name.String())
		assert.Equal(t, *tdPut.Description, trustDomain.Description)
	})
}
