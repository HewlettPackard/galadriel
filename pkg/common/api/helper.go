package api

import (
	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (td TrustDomain) ToEntity() (*entity.TrustDomain, error) {
	harvesterSpiffeID, err := spiffeid.FromString(*td.HarvesterSpiffeId)
	if err != nil {
		return nil, common.ErrWrongSPIFFEID{Cause: err}
	}

	tdName, err := spiffeid.TrustDomainFromString(td.Name)
	if err != nil {
		return nil, common.ErrWrongTrustDomain{Cause: err}
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
