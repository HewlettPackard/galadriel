package datastore

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/gca/api"
	"github.com/google/uuid"
	pgt "github.com/jackc/pgtype"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func toPGType(uuid uuid.UUID) (pgt.UUID, error) {

	pgID := pgt.UUID{}
	err := pgID.Set(uuid)
	if err != nil {
		return pgt.UUID{}, fmt.Errorf("failed to convert to PGTYPE")
	}
	return pgID, err
}

func (td TrustDomain) ToEntity() (*api.TrustDomain, error) {
	trustDomain, err := spiffeid.TrustDomainFromString(td.Name)
	if err != nil {
		return nil, err
	}

	id := uuid.NullUUID{
		UUID:  td.ID.Bytes,
		Valid: true,
	}

	result := &api.TrustDomain{
		ID:               id,
		Name:             trustDomain,
		OnboardingBundle: td.OnboardingBundle,
		CreatedAt:        td.CreatedAt,
		UpdatedAt:        td.UpdatedAt,
	}

	if td.Description.Valid {
		result.Description = td.Description.String
	}

	if td.HarvesterSpiffeID.Valid {
		id, err := spiffeid.FromStringf(td.HarvesterSpiffeID.String)
		if err != nil {
			return nil, fmt.Errorf("cannot convert model to entity: %w", err)
		}
		result.HarvesterSpiffeID = id
	}

	return result, nil
}

func (jt JoinToken) ToEntity() (*api.JoinToken, error) {
	id := uuid.NullUUID{
		UUID:  jt.ID.Bytes,
		Valid: true,
	}

	result := &api.JoinToken{
		ID:            id,
		Token:         jt.Token,
		ExpiresAt:     jt.ExpiresAt,
		Used:          jt.Used,
		TrustDomainID: jt.TrustDomainID.Bytes,
		CreatedAt:     jt.CreatedAt,
		UpdatedAt:     jt.UpdatedAt,
	}
	return result, nil
}
