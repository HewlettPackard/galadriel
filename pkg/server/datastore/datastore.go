package datastore

import (
	"context"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"

	"github.com/google/uuid"
)

// DataStore defines the data storage interface.
type DataStore interface {
	CreateMember(ctx context.Context, member *entity.Member) (*entity.Member, error)
	CreateMembership(ctx context.Context, membership *entity.Membership) (*entity.Membership, error)
	CreateRelationship(ctx context.Context, relationship *entity.Relationship) (*entity.Relationship, error)
	GetMemberByID(ctx context.Context, memberID uuid.UUID) (*entity.Member, error)
}
