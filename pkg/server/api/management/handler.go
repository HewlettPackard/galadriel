package management

import (
	"github.com/labstack/echo/v4"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

// (GET /spireServers)
func (server Server) GetSpireServers(ctx echo.Context, params GetSpireServersParams) error {
	return nil
}

// (POST /spireServers)
func (server Server) CreateSpireServer(ctx echo.Context) error {
	return nil
}

// (DELETE /spireServers/{spireServerId})
func (server Server) DeleteSpireServer(ctx echo.Context, spireServerId int64) error {
	return nil
}

// (PUT /spireServers/{spireServerId})
func (server Server) UpdateSpireServer(ctx echo.Context, spireServerId int64) error {
	return nil
}

// (PUT /trustBundles/{trustBundleId})
func (server Server) UpdateTrustBundle(ctx echo.Context, trustBundleId int64) error {
	return nil
}

// GetFederationGroupMembershipsParams defines parameters for GetFederationGroupMemberships.
func (server Server) GetFederationGroupMemberships(ctx echo.Context, params GetFederationGroupMembershipsParams) error {
	return nil
}

// (POST /federationGroupMemberships)
func (server Server) CreateFederationGroupMembership(ctx echo.Context) error {
	return nil
}

// (DELETE /federationGroupMemberships/{membershipID})
func (server Server) DeletefederationGroupMembership(ctx echo.Context, membershipID int64) error {
	return nil
}

// (GET /federationGroupMemberships/{membershipID})
func (server Server) GetFederationGroupMembershipbyID(ctx echo.Context, membershipID int64) error {
	return nil
}

// (PUT /federationGroupMemberships/{membershipID})
func (server Server) UpdatefederationGroupMembership(ctx echo.Context, membershipID int64) error {
	return nil
}

// (GET /federationGroups)
func (server Server) GetFederationGroups(ctx echo.Context, params GetFederationGroupsParams) error {
	return nil
}

// (POST /federationGroups)
func (server Server) CreateFederationGroup(ctx echo.Context) error {
	return nil
}

// (DELETE /federationGroups/{federationGroupID})
func (server Server) DeletefederationGroup(ctx echo.Context, federationGroupID int64) error {
	return nil
}

// (GET /federationGroups/{federationGroupID})
func (server Server) GetFederationGroupbyID(ctx echo.Context, federationGroupID int64) error {
	return nil
}

// (PUT /federationGroups/{federationGroupID})
func (server Server) UpdatefederationGroup(ctx echo.Context, federationGroupID int64) error {
	return nil
}

// (GET /federationRelationships)
func (server Server) GetFederationRelationships(ctx echo.Context, params GetFederationRelationshipsParams) error {
	return nil
}

// (POST /federationRelationships)
func (server Server) CreateFederationRelationship(ctx echo.Context) error {
	return nil
}

// (GET /federationRelationships/{relationshipID})
func (server Server) GetFederationRelationshipbyID(ctx echo.Context, relationshipID int64) error {
	return nil
}

// (PUT /federationRelationships/{relationshipID})
func (server Server) UpdateFederationRelationshipship(ctx echo.Context, relationshipID int64) error {
	return nil
}

// (GET /organizations)
func (server Server) GetOrganizations(ctx echo.Context, params GetOrganizationsParams) error {
	return nil
}

// (POST /organizations)
func (server Server) CreateOrganization(ctx echo.Context) error {
	return nil
}

// (DELETE /organizations/{orgID})
func (server Server) DeleteOrganization(ctx echo.Context, orgID int64) error {
	return nil
}

// (GET /organizations/{orgID})
func (server Server) GetOrgbyID(ctx echo.Context, orgID int64) error {
	return nil
}

// (PUT /organizations/{orgID})
func (server Server) UpdateOrganizaion(ctx echo.Context, orgID int64) error {
	return nil
}
