package db

import (
	"github.com/HewlettPackard/galadriel/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (fg FederationGroup) ToEntity() (*entity.FederationGroup, error) {
	status, err := fg.Status.ToEntity()
	if err != nil {
		return nil, err
	}

	id := uuid.NullUUID{
		UUID:  fg.ID.Bytes,
		Valid: true,
	}
	return &entity.FederationGroup{
		ID:          id,
		Name:        fg.Name,
		Description: fg.Description.String,
		Status:      status,
		CreatedAt:   fg.CreatedAt,
		UpdatedAt:   fg.UpdatedAt,
	}, nil
}

func (m Member) ToEntity() (*entity.Member, error) {
	td, err := spiffeid.TrustDomainFromString(m.TrustDomain)
	if err != nil {
		return nil, err
	}

	status, err := m.Status.ToEntity()
	if err != nil {
		return nil, err
	}

	id := uuid.NullUUID{
		UUID:  m.ID.Bytes,
		Valid: true,
	}
	return &entity.Member{
		ID:          id,
		TrustDomain: td,
		Status:      status,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

func (m Membership) ToEntity() (*entity.Membership, error) {
	status, err := m.Status.ToEntity()
	if err != nil {
		return nil, err
	}

	id := uuid.NullUUID{
		UUID:  m.ID.Bytes,
		Valid: true,
	}

	return &entity.Membership{
		ID:                id,
		MemberID:          m.MemberID.Bytes,
		FederationGroupID: m.FederationGroupID.Bytes,
		Status:            status,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}, nil
}

func (b Bundle) ToEntity() (*entity.Bundle, error) {
	id := uuid.NullUUID{
		UUID:  b.ID.Bytes,
		Valid: true,
	}

	var pem string
	if b.SvidPem.Valid {
		pem = b.SvidPem.String
	}

	return &entity.Bundle{
		ID:           id,
		RawBundle:    b.RawBundle,
		Digest:       b.Digest,
		SignedBundle: b.SignedBundle,
		TlogID:       b.TlogID,
		SvidPem:      pem,
		MemberID:     b.MemberID.Bytes,
		CreatedAt:    b.CreatedAt,
		UpdatedAt:    b.UpdatedAt,
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
		ID:        id,
		Token:     t.Token,
		Expiry:    t.Expiry,
		Used:      used,
		MemberID:  t.MemberID.Bytes,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func (h Harvester) ToEntity() *entity.Harvester {
	id := uuid.NullUUID{
		UUID:  h.ID.Bytes,
		Valid: true,
	}

	isLeader := false
	if h.IsLeader.Valid {
		isLeader = h.IsLeader.Bool
	}

	return &entity.Harvester{
		ID:          id,
		MemberID:    h.MemberID.Bytes,
		IsLeader:    isLeader,
		LeaderUntil: h.LeaderUntil,
		CreatedAt:   h.CreatedAt,
		UpdatedAt:   h.UpdatedAt,
	}

}

func (s *Status) ToEntity() (entity.Status, error) {
	switch *s {
	case StatusPending:
		return entity.StatusPending, nil
	case StatusActive:
		return entity.StatusActive, nil
	case StatusDisabled:
		return entity.StatusDisabled, nil
	case StatusDenied:
		return entity.StatusDenied, nil
	default:
		return "", errors.Errorf("cannot map model status to entity: %v", *s)
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
