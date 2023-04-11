package harvester

import (
	"fmt"

	common "github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (b BundlePut) ToEntity() (*entity.Bundle, error) {
	td, err := spiffeid.TrustDomainFromString(b.TrustDomain)
	if err != nil {
		return nil, fmt.Errorf("malformed trust domain[%v]: %w", b.TrustDomain, err)
	}

	return &entity.Bundle{
		Data:            []byte(b.TrustBundle),
		Signature:       []byte(b.Signature),
		TrustDomainName: td,
		// TODO: do we need to store it in PEM or DER?
		SigningCertificate: []byte(b.SigningCertificate),
	}, nil
}

func RelationshipFromEntity(r *entity.Relationship) *common.Relationship {
	return &common.Relationship{
		Id:                  r.ID.UUID,
		TrustDomainAId:      r.TrustDomainAID,
		TrustDomainBId:      r.TrustDomainBID,
		TrustDomainAConsent: r.TrustDomainAConsent,
		TrustDomainBConsent: r.TrustDomainBConsent,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
	}
}
