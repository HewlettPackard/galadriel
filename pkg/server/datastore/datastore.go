package datastore

import (
	"context"

	entity "github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/api/management"
	"github.com/spiffe/spire/proto/spire/common"
)

// DataStore defines the data storage interface.
type DataStore interface {
	CreateMember(context.Context, *entity.Member) (*entity.Member, error)
	CreateMembership(ctx context.Context, membership *entity.Membership) (*entity.Membership, error)
	CreateTrustBundle(ctx context.Context, trust *common.Bundle, memberID uint) (*common.Bundle, error)
	CreateRelationship(ctx context.Context, relationship *entity.Relationship, sourceID uint, targetID uint) (*entity.Relationship, error)
	RetrieveMembershipsByID(ctx context.Context, memberID uint) (*[]management.FederationGroupMembership, error)
	RetrieveRelationshipsByID(ctx context.Context, memberID uint) (*[]entity.Relationship, error)
	RetrieveTrustBundlesByID(ctx context.Context, memberID uint) (*[]common.Bundle, error)
	RetrieveMemberByID(context.Context, entity.Member) (*entity.Member, error)
	RetrieveMembershipByID(ctx context.Context, membershipID uint) (*management.FederationGroupMembership, error)
	RetrieveRelationshipBySourceandTargetID(ctx context.Context, source uint, target uint) (*entity.Relationship, error)
	RetrieveTrustbundleByMemberID(ctx context.Context, memberID string) (*common.Bundle, error)
	UpdateMember(context.Context, management.SpireServer) error
	UpdateMembership(context.Context, management.FederationGroupMembership) error
	UpdateTrust(context.Context, common.Bundle) error
	DeleteOrganizationByID(ctx context.Context, orgID uint) error
	DeleteMemberbyID(ctx context.Context, memberID uint) error
	DeleteMembershipsByID(ctx context.Context, memberid uint) error
	DeleteRelationshipsByID(ctx context.Context, memberid uint) error
	DeleteTrustbundlesByID(ctx context.Context, memberid uint) error
	DeleteMembershipByID(ctx context.Context, membershipID uint) error
	DeleteRelationshipBySourceTargetID(ctx context.Context, source uint, target uint) error
	DeleteTrustBundleByID(ctx context.Context, memberID string) error
}