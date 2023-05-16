package api

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (td TrustDomain) ToEntity() (*entity.TrustDomain, error) {

	var harvesterSpiffeID spiffeid.ID
	if td.HarvesterSpiffeId != nil {
		hSID, err := spiffeid.FromString(*td.HarvesterSpiffeId)
		if err != nil {
			return nil, fmt.Errorf("malformed SPIFFE ID[%v]: %w", *td.HarvesterSpiffeId, err)
		}

		harvesterSpiffeID = hSID
	}

	tdName, err := spiffeid.TrustDomainFromString(td.Name)
	if err != nil {
		return nil, fmt.Errorf("malformed trust domain[%v]: %w", td.Name, err)
	}

	description := ""
	if td.Description != nil {
		description = *td.Description
	}

	onboardingBundle := []byte{}
	if td.OnboardingBundle != nil {
		onboardingBundle = []byte(*td.OnboardingBundle)
	}

	uuid := uuid.NullUUID{
		UUID:  td.Id,
		Valid: true,
	}

	return &entity.TrustDomain{
		ID:                uuid,
		Name:              tdName,
		CreatedAt:         td.CreatedAt,
		UpdatedAt:         td.UpdatedAt,
		Description:       description,
		OnboardingBundle:  onboardingBundle,
		HarvesterSpiffeID: harvesterSpiffeID,
	}, nil
}

func TrustDomainFromEntity(entity *entity.TrustDomain) *TrustDomain {
	onboardingBundle := string(entity.OnboardingBundle)
	harvesterSpiffeID := entity.HarvesterSpiffeID.String()

	return &TrustDomain{
		Id:                entity.ID.UUID,
		Name:              entity.Name.String(),
		Description:       &entity.Description,
		UpdatedAt:         entity.UpdatedAt,
		CreatedAt:         entity.CreatedAt,
		OnboardingBundle:  &onboardingBundle,
		HarvesterSpiffeId: &harvesterSpiffeID,
	}
}

func (r Relationship) ToEntity() (*entity.Relationship, error) {
	var id uuid.NullUUID
	if r.Id != uuid.Nil {
		id = uuid.NullUUID{UUID: r.Id, Valid: true}
	}

	tdAName, err := spiffeid.TrustDomainFromString(*r.TrustDomainAName)
	if err != nil {
		return nil, fmt.Errorf("malformed trust domain[%v]: %w", r.TrustDomainAName, err)
	}

	tdBName, err := spiffeid.TrustDomainFromString(*r.TrustDomainBName)
	if err != nil {
		return nil, fmt.Errorf("malformed trust domain[%v]: %w", r.TrustDomainBName, err)
	}

	return &entity.Relationship{
		ID:                  id,
		TrustDomainAID:      r.TrustDomainAId,
		TrustDomainBID:      r.TrustDomainBId,
		TrustDomainAName:    tdAName,
		TrustDomainBName:    tdBName,
		TrustDomainAConsent: entity.ConsentStatus(r.TrustDomainAConsent),
		TrustDomainBConsent: entity.ConsentStatus(r.TrustDomainBConsent),
	}, nil
}

func RelationshipFromEntity(entity *entity.Relationship) *Relationship {
	trustDomainAName := entity.TrustDomainAName.String()
	trustDomainBName := entity.TrustDomainBName.String()

	return &Relationship{
		Id:                  entity.ID.UUID,
		TrustDomainAId:      entity.TrustDomainAID,
		TrustDomainBId:      entity.TrustDomainBID,
		TrustDomainAName:    &trustDomainAName,
		TrustDomainBName:    &trustDomainBName,
		TrustDomainAConsent: ConsentStatus(entity.TrustDomainAConsent),
		TrustDomainBConsent: ConsentStatus(entity.TrustDomainBConsent),
		CreatedAt:           entity.CreatedAt,
		UpdatedAt:           entity.UpdatedAt,
	}
}

// FilterRelationships filters a slice of Relationship entities based on a trust domain ID and consent status.
// If the trust domain ID is nil, it filters based on the consent status only.
// If the trust domain ID is not nil, it filters based on both the trust domain ID and the consent status.
// Parameters:
// - trustDomain: The ID of the trust domain for which relationships are to be filtered. Can be nil.
// - relationships: The slice of Relationship entities to be filtered.
// - status: The consent status to be used as a filter criterion.
// Return: A slice of Relationship entities that match the filter criteria.
func FilterRelationships(trustDomain *uuid.UUID, relationships []*entity.Relationship, status ConsentStatus) []*entity.Relationship {
	filtered := make([]*entity.Relationship, 0, len(relationships))

	for _, relationship := range relationships {
		trustDomainA := relationship.TrustDomainAID
		trustDomainB := relationship.TrustDomainBID
		trustDomainAConsent := ConsentStatus(relationship.TrustDomainAConsent)
		trustDomainBConsent := ConsentStatus(relationship.TrustDomainBConsent)

		var isConsentStatusMatch bool
		if trustDomain == nil {
			isConsentStatusMatch = trustDomainAConsent == status || trustDomainBConsent == status
		} else {
			isConsentStatusMatch = (trustDomainA == *trustDomain && trustDomainAConsent == status) ||
				(trustDomainB == *trustDomain && trustDomainBConsent == status)
		}

		if isConsentStatusMatch {
			filtered = append(filtered, relationship)
		}
	}

	// Trim the slice to the actual length to free up unused capacity
	return filtered[:]
}

// MapRelationships transforms a slice of Relationship entities to a slice of API Relationship representations.
// Parameters:
// - relationships: A slice of Relationship entities to be transformed.
// Return: A slice of API Relationship representations.
func MapRelationships(relationships ...*entity.Relationship) []*Relationship {
	cRelationships := make([]*Relationship, len(relationships))

	for i, r := range relationships {
		cRelationships[i] = RelationshipFromEntity(r)
	}

	return cRelationships
}
