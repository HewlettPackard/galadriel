package datastore

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (m TrustDomain) ToEntity() (*entity.TrustDomain, error) {
	td, err := spiffeid.TrustDomainFromString(m.Name)
	if err != nil {
		return nil, err
	}

	id := uuid.NullUUID{
		UUID:  m.ID.Bytes,
		Valid: true,
	}

	result := &entity.TrustDomain{
		ID:               id,
		Name:             td,
		OnboardingBundle: m.OnboardingBundle,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}

	if m.Description.Valid {
		result.Description = m.Description.String
	}

	if m.HarvesterSpiffeID.Valid {
		id, err := spiffeid.FromStringf(m.HarvesterSpiffeID.String)
		if err != nil {
			return nil, fmt.Errorf("cannot convert model to entity: %v", err)
		}
		result.HarvesterSpiffeID = id
	}

	return result, nil
}

func (m Relationship) ToEntity() (*entity.Relationship, error) {
	id := uuid.NullUUID{
		UUID:  m.ID.Bytes,
		Valid: true,
	}

	return &entity.Relationship{
		ID:                  id,
		TrustDomainAID:      m.TrustDomainAID.Bytes,
		TrustDomainBID:      m.TrustDomainBID.Bytes,
		TrustDomainAConsent: m.TrustDomainAConsent,
		TrustDomainBConsent: m.TrustDomainBConsent,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
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
		DigestAlgorithm:    b.DigestAlgorithm,
		SignatureAlgorithm: b.SignatureAlgorithm,
		SigningCert:        b.SigningCert,
		TrustDomainID:      b.TrustDomainID.Bytes,
		CreatedAt:          b.CreatedAt,
		UpdatedAt:          b.UpdatedAt,
	}, nil
}

func (t JoinToken) ToEntity() *entity.JoinToken {
	id := uuid.NullUUID{
		UUID:  t.ID.Bytes,
		Valid: true,
	}

	used := false
	if t.Used.Valid {
		used = t.Used.Bool
	}

	return &entity.JoinToken{
		ID:            id,
		Token:         t.Token,
		ExpiresAt:     t.ExpiresAt,
		Used:          used,
		TrustDomainID: t.TrustDomainID.Bytes,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
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
