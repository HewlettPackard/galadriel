package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFilterRelationships(t *testing.T) {
	relationships := []*Relationship{
		{
			ID:                  uuid.NullUUID{UUID: uuid.New(), Valid: true},
			TrustDomainAID:      uuid.New(),
			TrustDomainBID:      uuid.New(),
			TrustDomainAConsent: "approved",
			TrustDomainBConsent: "denied",
		},
		{
			ID:                  uuid.NullUUID{UUID: uuid.New(), Valid: true},
			TrustDomainAID:      uuid.New(),
			TrustDomainBID:      uuid.New(),
			TrustDomainAConsent: "denied",
			TrustDomainBConsent: "approved",
		},
	}

	trustDomain := relationships[0].TrustDomainAID
	status := ConsentStatus("approved")

	// Call FilterRelationships
	filtered := FilterRelationships(relationships, status, &trustDomain)

	assert.Equal(t, 1, len(filtered))
	assert.Equal(t, relationships[0].ID.UUID, filtered[0].ID.UUID)
	assert.Equal(t, relationships[0].TrustDomainAID, filtered[0].TrustDomainAID)
	assert.Equal(t, relationships[0].TrustDomainBID, filtered[0].TrustDomainBID)
	assert.Equal(t, relationships[0].TrustDomainAConsent, filtered[0].TrustDomainAConsent)
	assert.Equal(t, relationships[0].TrustDomainBConsent, filtered[0].TrustDomainBConsent)
}
