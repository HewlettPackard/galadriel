package entity

import (
	"github.com/google/uuid"
)

// FilterRelationships filters a slice of Relationship entities based on a trust domain ID and consent status.
// If the trust domain ID is nil, it filters based on the consent status only.
// If the trust domain ID is not nil, it filters based on both the trust domain ID and the consent status.
// Parameters:
// - relationships: The slice of Relationship entities to be filtered.
// - status: The consent status to be used as a filter criterion.
// - trustDomain: The ID of the trust domain for which relationships are to be filtered. Can be nil.
// Return: A slice of Relationship entities that match the filter criteria.
func FilterRelationships(relationships []*Relationship, status ConsentStatus, trustDomain *uuid.UUID) []*Relationship {
	// Pre-allocate space for the slice to avoid unnecessary allocations
	filtered := make([]*Relationship, 0, len(relationships))

	for _, relationship := range relationships {
		trustDomainA, trustDomainB := relationship.TrustDomainAID, relationship.TrustDomainBID
		trustDomainAConsent, trustDomainBConsent := relationship.TrustDomainAConsent, relationship.TrustDomainBConsent

		var isConsentStatusMatch bool
		if trustDomain != nil {
			// If the trust domain ID is not nil, check if it matches either of the two trust domain IDs
			// in the relationship along with the consent status
			isConsentStatusMatch = (trustDomainA == *trustDomain && trustDomainAConsent == status) ||
				(trustDomainB == *trustDomain && trustDomainBConsent == status)
		} else {
			isConsentStatusMatch = trustDomainAConsent == status || trustDomainBConsent == status
		}

		if isConsentStatusMatch {
			filtered = append(filtered, relationship)
		}
	}

	// Trim the slice to the actual length to free up unused capacity
	return filtered[:]
}
