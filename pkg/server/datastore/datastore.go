package datastore

import (
	"context"
	"net/url"
	"time"

	management "github.com/HewlettPackard/galadriel/pkg/server/api/management"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/spire/proto/spire/common"
)

// DataStore defines the data storage interface.
type DataStore interface {
	CreateMember(context.Context, *management.SpireServer) (*management.SpireServer, error)
	CreateMembership(ctx context.Context, membership *management.FederationGroupMembership, memberID uint, bridgeID uint) (*management.FederationGroupMembership, error)
	CreateTrustBundle(ctx context.Context, trust *common.Bundle, memberID uint) (*common.Bundle, error)
	CreateRelationship(ctx context.Context, newrelation *FederationRelationship, sourceID uint, targetID uint) (*FederationRelationship, error)
	RetrieveAllMembershipsByMemberID(ctx context.Context, memberID uint) (*[]management.FederationGroupMembership, error)
	RetrieveAllRelationshipsByMemberID(ctx context.Context, memberID uint) (*[]FederationRelationship, error)
	RetrieveAllTrustBundlesByMemberID(ctx context.Context, memberID uint) (*[]common.Bundle, error)
	RetrieveMemberByID(ctx context.Context, memberID uint) (*management.SpireServer, error)
	RetrieveMembershipByCreationDate(ctx context.Context, date time.Time) (*management.FederationGroupMembership, error)
	RetrieveMembershipByToken(ctx context.Context, token string) (*management.FederationGroupMembership, error)
	RetrieveRelationshipBySourceandTargetID(ctx context.Context, source uint, target uint) (*FederationRelationship, error)
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
type BundleEndpointType string

type FederationRelationship struct {
	TrustDomain           spiffeid.TrustDomain
	BundleEndpointURL     *url.URL
	BundleEndpointProfile BundleEndpointType
	TrustDomainBundle     *common.Bundle

	// Fields only used for 'https_spiffe' bundle endpoint profile
	EndpointSPIFFEID spiffeid.ID
}
