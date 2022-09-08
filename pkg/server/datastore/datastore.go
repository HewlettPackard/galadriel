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
	CreateRelationship(ctx context.Context, relationship *entity.Relationship) (*entity.Relationship, error)
	RetrieveMemberByID(ctx context.Context, memberID uuid.UUID) (*entity.Member, error)
	RetrieveMembershipByID(ctx context.Context, membershipID uuid.UUID) (*entity.Membership, error)
	RetrieveRelationshipByID(ctx context.Context, relationshipID uuid.UUID) (*entity.Relationship, error)
	UpdateMember(ctx context.Context, member *entity.Member) (*entity.Member, error)
	UpdateMembership(ctx context.Context, membership *entity.Membership) (*entity.Membership, error)
	UpdateRelationship(ctx context.Context, relationship *entity.Relationship) (*common.Bundle, error)
	DeleteMemberByID(ctx context.Context, memberID uuid.UUID) error
	DeleteMembershipByID(ctx context.Context, membershipID uuid.UUID) error
	DeleteRelationshipByID(ctx context.Context, relationshipID uuid.UUID) error
}
