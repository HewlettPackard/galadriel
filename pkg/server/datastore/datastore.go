package datastore

import (
	"context"
	"time"

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
	RetrieveAllMembershipsByMemberID(ctx context.Context, memberID uint) (*[]management.FederationGroupMembership, error)
	RetrieveAllRelationshipsByMemberID(ctx context.Context, memberID uint) (*[]entity.Relationship, error)
	RetrieveAllTrustBundlesByMemberID(ctx context.Context, memberID uint) (*[]common.Bundle, error)
	RetrieveMemberByID(context.Context, entity.Member) (*entity.Member, error)
	RetrieveMembershipByCreationDate(ctx context.Context, date time.Time) (*management.FederationGroupMembership, error)
	RetrieveMembershipByToken(ctx context.Context, token string) (*management.FederationGroupMembership, error)
	RetrieveRelationshipBySourceandTargetID(ctx context.Context, source uint, target uint) (*entity.Relationship, error)
	RetrieveTrustbundleByMemberID(ctx context.Context, memberID string) (*common.Bundle, error)
	UpdateMember(context.Context, management.SpireServer) error
	UpdateMembership(context.Context, management.FederationGroupMembership) error
	UpdateTrust(context.Context, common.Bundle) error
	DeleteOrganizationByID(ctx context.Context, orgID uint) error
	DeleteMemberbyID(ctx context.Context, memberID uint) error
	DeleteAllMembershipsByMemberID(ctx context.Context, memberid uint) error
	DeleteAllMembershipsByBridgeID(ctx context.Context, bridgeid uint) error
	DeleteAllRelationshipsByMemberID(ctx context.Context, memberid uint) error
	DeleteAllTrustbundlesByMemberID(ctx context.Context, memberid uint) error
	DeleteMembershipByToken(ctx context.Context, name string) error
	DeleteRelationshipBySourceTargetID(ctx context.Context, source uint, target uint) error
	DeleteTrustBundleByMemberID(ctx context.Context, memberID string) error
}
