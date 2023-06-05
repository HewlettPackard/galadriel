package sqlite

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (td TrustDomain) ToEntity() (*entity.TrustDomain, error) {
	trustDomain, err := spiffeid.TrustDomainFromString(td.Name)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(td.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot convert model to entity: %v", err)
	}
	nullID := uuid.NullUUID{
		UUID:  id,
		Valid: true,
	}

	result := &entity.TrustDomain{
		ID:        nullID,
		Name:      trustDomain,
		CreatedAt: td.CreatedAt,
		UpdatedAt: td.UpdatedAt,
	}

	if td.Description.Valid {
		result.Description = td.Description.String
	}

	return result, nil
}

func (r Relationship) ToEntity() (*entity.Relationship, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot convert model to entity: %v", err)
	}
	nullID := uuid.NullUUID{
		UUID:  id,
		Valid: true,
	}

	tdAID, err := uuid.Parse(r.TrustDomainAID)
	if err != nil {
		return nil, fmt.Errorf("cannot convert model to entity: %v", err)
	}
	tdBID, err := uuid.Parse(r.TrustDomainBID)
	if err != nil {
		return nil, fmt.Errorf("cannot convert model to entity: %v", err)
	}

	return &entity.Relationship{
		ID:                  nullID,
		TrustDomainAID:      tdAID,
		TrustDomainBID:      tdBID,
		TrustDomainAConsent: entity.ConsentStatus(r.TrustDomainAConsent),
		TrustDomainBConsent: entity.ConsentStatus(r.TrustDomainBConsent),
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
	}, nil
}

func (b Bundle) ToEntity() (*entity.Bundle, error) {
	id, err := uuid.Parse(b.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot convert model to entity: %v", err)
	}
	nullID := uuid.NullUUID{
		UUID:  id,
		Valid: true,
	}

	tdID, err := uuid.Parse(b.TrustDomainID)
	if err != nil {
		return nil, fmt.Errorf("cannot convert model to entity: %v", err)
	}

	return &entity.Bundle{
		ID:                 nullID,
		Data:               b.Data,
		Digest:             b.Digest,
		Signature:          b.Signature,
		SigningCertificate: b.SigningCertificate,
		TrustDomainID:      tdID,
		CreatedAt:          b.CreatedAt,
		UpdatedAt:          b.UpdatedAt,
	}, nil
}

func (jt JoinToken) ToEntity() (*entity.JoinToken, error) {
	id, err := uuid.Parse(jt.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot convert model to entity: %v", err)
	}
	nullID := uuid.NullUUID{
		UUID:  id,
		Valid: true,
	}

	tdID, err := uuid.Parse(jt.TrustDomainID)
	if err != nil {
		return nil, fmt.Errorf("cannot convert model to entity: %v", err)
	}

	return &entity.JoinToken{
		ID:            nullID,
		Token:         jt.Token,
		ExpiresAt:     jt.ExpiresAt,
		Used:          jt.Used,
		TrustDomainID: tdID,
		CreatedAt:     jt.CreatedAt,
		UpdatedAt:     jt.UpdatedAt,
	}, nil
}
