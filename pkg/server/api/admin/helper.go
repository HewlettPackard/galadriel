package admin

import (
	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (r RelationshipRequest) ToEntity() *entity.Relationship {
	return &entity.Relationship{
		TrustDomainAID: r.TrustDomainAId,
		TrustDomainBID: r.TrustDomainBId,
	}
}

func (td TrustDomainPut) ToEntity() (*entity.TrustDomain, error) {
	tdName, err := spiffeid.TrustDomainFromString(td.Name)
	if err != nil {
		return nil, common.ErrWrongTrustDomain{Cause: err}
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
