package postgres

import (
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (td TrustDomain) ToEntity() (*entity.TrustDomain, error) {
	trustDomain, err := spiffeid.TrustDomainFromString(td.Name)
	if err != nil {
		return nil, err
	}

	id := uuid.NullUUID{
		UUID:  td.ID.Bytes,
		Valid: true,
	}

	result := &entity.TrustDomain{
		ID:        id,
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
	id := uuid.NullUUID{
		UUID:  r.ID.Bytes,
		Valid: true,
	}

	return &entity.Relationship{
		ID:                  id,
		TrustDomainAID:      r.TrustDomainAID.Bytes,
		TrustDomainBID:      r.TrustDomainBID.Bytes,
		TrustDomainAConsent: entity.ConsentStatus(r.TrustDomainAConsent),
		TrustDomainBConsent: entity.ConsentStatus(r.TrustDomainBConsent),
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
	}, nil
}

func (b Bundle) ToEntity() (*entity.Bundle, error) {
	id := uuid.NullUUID{
		UUID:  b.ID.Bytes,
		Valid: true,
	}

	return &entity.Bundle{
		ID:                 id,
		Data:               b.Data,
		Digest:             b.Digest,
		Signature:          b.Signature,
		SigningCertificate: b.SigningCertificate,
		TrustDomainID:      b.TrustDomainID.Bytes,
		CreatedAt:          b.CreatedAt,
		UpdatedAt:          b.UpdatedAt,
	}, nil
}

func (jt JoinToken) ToEntity() *entity.JoinToken {
	id := uuid.NullUUID{
		UUID:  jt.ID.Bytes,
		Valid: true,
	}

	return &entity.JoinToken{
		ID:            id,
		Token:         jt.Token,
		ExpiresAt:     jt.ExpiresAt,
		Used:          jt.Used,
		TrustDomainID: jt.TrustDomainID.Bytes,
		CreatedAt:     jt.CreatedAt,
		UpdatedAt:     jt.UpdatedAt,
	}
}

func uuidToPgType(id uuid.UUID) (pgtype.UUID, error) {
	pgID := pgtype.UUID{}
	err := pgID.Set(id)
	if err != nil {
		return pgtype.UUID{}, errors.Errorf("failed converting UUID to Postgres UUID type: %v", err)
	}
	return pgID, err
}
