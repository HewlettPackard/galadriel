package admin

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (r *PutRelationshipRequest) ToEntity() (*entity.Relationship, error) {
	tdA, err := spiffeid.TrustDomainFromString(r.TrustDomainAName)
	if err != nil {
		return nil, fmt.Errorf("malformed trust domain[%q]: %v", r.TrustDomainAName, err)
	}

	tdB, err := spiffeid.TrustDomainFromString(r.TrustDomainBName)
	if err != nil {
		return nil, fmt.Errorf("malformed trust domain[%q]: %v", r.TrustDomainBName, err)
	}

	return &entity.Relationship{
		TrustDomainAName: tdA,
		TrustDomainBName: tdB,
	}, nil
}

func (td *PutTrustDomainRequest) ToEntity() (*entity.TrustDomain, error) {
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
