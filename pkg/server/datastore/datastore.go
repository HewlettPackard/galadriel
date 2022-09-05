package datastore

import (
	"context"

	entity "github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/spiffe/spire/proto/spire/common"
)

// DataStore defines the data storage interface.
type DataStore interface {
	CreateMember(ctx context.Context, member *entity.Member) (*entity.Member, error)
	CreateMembership(ctx context.Context, membership *entity.Membership) (*entity.Membership, error)
	CreateTrustBundle(ctx context.Context, trustBundle *common.Bundle, memberID uuid.UUID) (*common.Bundle, error)
	CreateRelationship(ctx context.Context, relationship *entity.Relationship, sourceMemberID uuid.UUID, targetMemberID uuid.UUID) (*entity.Relationship, error)
	RetrieveMemberByID(ctx context.Context, memberID uuid.UUID) (*entity.Member, error)
	RetrieveMembershipByID(ctx context.Context, membershipID uuid.UUID) (*entity.Membership, error)
	RetrieveRelationshipBySourceandTargetID(ctx context.Context, sourceMemberID uuid.UUID, targetMemberID uuid.UUID) (*entity.Relationship, error)
	RetrieveTrustBundleByID(ctx context.Context, trustID uuid.UUID) (*common.Bundle, error)
	UpdateMember(ctx context.Context, member *entity.Member) (*entity.Member, error)
	UpdateMembership(ctx context.Context, membership *entity.Membership) (*entity.Membership, error)
	UpdateRelationship(ctx context.Context, relationship *entity.Relationship) (*common.Bundle, error)
	UpdateTrust(ctx context.Context, trustbundle *common.Bundle) (*common.Bundle, error)
	DeleteMemberbyID(ctx context.Context, memberID uuid.UUID) error
	DeleteMembershipByID(ctx context.Context, membershipID uuid.UUID) error
	DeleteRelationshipBySourceTargetID(ctx context.Context, sourceMemberID uuid.UUID, targetMemberID uuid.UUID) error
	DeleteTrustBundleByID(ctx context.Context, memberID uuid.UUID) error
}
