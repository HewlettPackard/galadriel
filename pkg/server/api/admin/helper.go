package admin

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (r RelationshipRequest) ToEntity() *entity.Relationship {
	return &entity.Relationship{
		TrustDomainAName: spiffeid.RequireTrustDomainFromString(r.TrustDomainAName),
		TrustDomainBName: spiffeid.RequireTrustDomainFromString(r.TrustDomainBName),
	}
}

func (td TrustDomainPut) ToEntity() (*entity.TrustDomain, error) {
	tdName, err := spiffeid.TrustDomainFromString(td.Name)
	if err != nil {
		return nil, fmt.Errorf("malformed trust domain[%v]: %w", td.Name, err)
	}

	description := ""
	if td.Description != nil {
		description = *td.Description
	}

	return &entity.TrustDomain{
		Name:        tdName,
		Description: description,
	}, nil
}
