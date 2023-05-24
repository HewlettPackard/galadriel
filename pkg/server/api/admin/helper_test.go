package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	td1 = "test.com"
	td2 = "test2.com"
)

func TestRelationshipRequestToEntity(t *testing.T) {
	t.Run("Full fill correctly the relationship entity model", func(t *testing.T) {
		releationshipRequest := PutRelationshipRequest{
			TrustDomainAName: td1,
			TrustDomainBName: td2,
		}

		r, err := releationshipRequest.ToEntity()
		assert.NoError(t, err)
		assert.NotNil(t, r)

		assert.Equal(t, releationshipRequest.TrustDomainAName, r.TrustDomainAName.String())
		assert.Equal(t, releationshipRequest.TrustDomainBName, r.TrustDomainBName.String())
	})
}

func TestTrustDomainPutToEntity(t *testing.T) {
	t.Run("Does not allow wrong trust domain names", func(t *testing.T) {
		description := "a cool description"

		tdPut := PutTrustDomainRequest{
			Name:        "A wrong trust domain name",
			Description: &description,
		}

		trustDomain, err := tdPut.ToEntity()
		assert.ErrorContains(t, err, "malformed trust domain[A wrong trust domain name]")
		assert.Nil(t, trustDomain)
	})

	t.Run("Full fill correctly the trust domain entity model", func(t *testing.T) {
		description := "a cool description"

		tdPut := PutTrustDomainRequest{
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
