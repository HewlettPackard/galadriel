package datastore

import (
	"context"
	"net/url"
	"time"

	management "github.com/HewlettPackard/galadriel/pkg/server/api/management"
	"github.com/labstack/echo/v4"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/spire/proto/spire/common"
)

// DataStore defines the data storage interface.
type DataStore interface {
	CreateMember(echo.Context, *management.SpireServer) (*management.SpireServer, error)
	CreateMembership(ctx context.Context, membership *management.FederationGroupMembership, memberID uint, bridgeID uint) (*management.FederationGroupMembership, error)
	CreateTrustBundle(ctx context.Context, trust *common.Bundle, memberID uint) (*common.Bundle, error)
	CreateRelationship(ctx context.Context, newrelation *FederationRelationship, sourceID uint, targetID uint) error
	RetrieveBridgebyID(ctx context.Context, brID uint) (*management.FederationGroup, error)
	RetrieveAllMembershipsbyBridgeID(ctx context.Context, bridgeID uint) (*[]management.FederationGroupMembership, error)
	RetrieveAllMembersbyBridgeID(ctx context.Context, bridgeID uint) (*[]management.SpireServer, error)
	RetrieveAllMembershipsbyMemberID(ctx context.Context, memberID uint) (*[]management.FederationGroupMembership, error)
	RetrieveAllRelationshipsbyMemberID(ctx context.Context, memberID uint) (*[]FederationRelationship, error)
	RetrieveAllTrustBundlesbyMemberID(ctx context.Context, memberID uint) (*[]common.Bundle, error)
	RetrieveMemberbyID(ctx context.Context, memberID uint) (*management.SpireServer, error)
	RetrieveMembershipbyCreationDate(ctx context.Context, date time.Time) (*management.FederationGroupMembership, error)
	RetrieveMembershipbyToken(ctx context.Context, token string) (*management.FederationGroupMembership, error)
	RetrieveRelationshipbySourceandTargetID(ctx context.Context, source uint, target uint) (*FederationRelationship, error)
	RetrieveTrustbundlebyMemberID(ctx context.Context, memberID string) (*common.Bundle, error)
	UpdateMember(context.Context, management.SpireServer) error
	UpdateMembership(context.Context, management.FederationGroupMembership) error
	UpdateTrust(context.Context, common.Bundle) error
	DeleteOrganizationbyID(ctx context.Context, orgID uint) error
	DeleteBridgebyID(ctx context.Context, bridgeID uint) error
	DeleteMemberbyID(ctx context.Context, memberID uint) error
	DeleteAllMembershipsbyMemberID(ctx context.Context, memberid uint) error
	DeleteAllMembershipsbyBridgeID(ctx context.Context, bridgeid uint) error
	DeleteAllRelationshipsbyMemberID(ctx context.Context, memberid uint) error
	DeleteAllTrustbundlesbyMemberID(ctx context.Context, memberid uint) error
	DeleteMembershipbyToken(ctx context.Context, name string) error
	DeleteRelationshipbySourceTargetID(ctx context.Context, source uint, target uint) error
	DeleteTrustBundlebyMemberID(ctx context.Context, memberID string) error
}
type BundleEndpointType string

type FederationRelationship struct {
	TrustDomain           spiffeid.TrustDomain
	BundleEndpointURL     *url.URL
	BundleEndpointProfile BundleEndpointType
	TrustDomainBundle     *common.Bundle

	// Fields only used for 'https_spiffe' bundle endpoint profile
	EndpointSPIFFEID spiffeid.ID
}
