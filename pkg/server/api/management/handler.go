package management

import (
	"github.com/labstack/echo/v4"
)

type Management struct {
}

func NewManagement() *Management {
	return &Management{}
}

// (GET /spireServers)
func (m *Management) GetSpireServers(ctx echo.Context, params GetSpireServersParams) error {
	return nil
}

// (POST /spireServers)
func (m *Management) CreateSpireServer(ctx echo.Context) error {
	return nil
}

// (DELETE /spireServers/{spireServerId})
func (m *Management) DeleteSpireServer(ctx echo.Context, spireServerId int64) error {
	return nil
}

// (PUT /spireServers/{spireServerId})
func (m *Management) UpdateSpireServer(ctx echo.Context, spireServerId int64) error {
	return nil
}

// (PUT /trustBundles/{trustBundleId})
func (m *Management) UpdateTrustBundle(ctx echo.Context, trustBundleId int64) error {
	return nil
}

// GetFederationGroupMembershipsParams defines parameters for GetFederationGroupMemberships.
func (m *Management) GetFederationGroupMemberships(ctx echo.Context, params GetFederationGroupMembershipsParams) error {
	return nil
}

// (POST /federationGroupMemberships)
func (m *Management) CreateFederationGroupMembership(ctx echo.Context) error {
	return nil
}

// (DELETE /federationGroupMemberships/{membershipID})
func (m *Management) DeleteFederationGroupMembership(ctx echo.Context, membershipID int64) error {
	return nil
}

// (GET /federationGroupMemberships/{membershipID})
func (m *Management) GetFederationGroupMembershipbyID(ctx echo.Context, membershipID int64) error {
	return nil
}

// (PUT /federationGroupMemberships/{membershipID})
func (m *Management) UpdateFederationGroupMembership(ctx echo.Context, membershipID int64) error {
	return nil
}

// (GET /federationGroups)
func (m *Management) GetFederationGroups(ctx echo.Context, params GetFederationGroupsParams) error {
	return nil
}

// (POST /federationGroups)
func (m *Management) CreateFederationGroup(ctx echo.Context) error {
	return nil
}

// (DELETE /federationGroups/{federationGroupID})
func (m *Management) DeleteFederationGroup(ctx echo.Context, federationGroupID int64) error {
	return nil
}

// (GET /federationGroups/{federationGroupID})
func (m *Management) GetFederationGroupbyID(ctx echo.Context, federationGroupID int64) error {
	return nil
}

// (PUT /federationGroups/{federationGroupID})
func (m *Management) UpdateFederationGroup(ctx echo.Context, federationGroupID int64) error {
	return nil
}

// (GET /federationRelationships)
func (m *Management) GetFederationRelationships(ctx echo.Context, params GetFederationRelationshipsParams) error {
	return nil
}

// (POST /federationRelationships)
func (m *Management) CreateFederationRelationship(ctx echo.Context) error {
	return nil
}

// (GET /federationRelationships/{relationshipID})
func (m *Management) GetFederationRelationshipbyID(ctx echo.Context, relationshipID int64) error {
	return nil
}

// (PUT /federationRelationships/{relationshipID})
func (m *Management) UpdateFederationRelationship(ctx echo.Context, relationshipID int64) error {
	return nil
}

// (GET /organizations)
func (m *Management) GetOrganizations(ctx echo.Context, params GetOrganizationsParams) error {
	return nil
}

// (POST /organizations)
func (m *Management) CreateOrganization(ctx echo.Context) error {
	return nil
}

// (DELETE /organizations/{orgID})
func (m *Management) DeleteOrganization(ctx echo.Context, orgID int64) error {
	return nil
}

// (GET /organizations/{orgID})
func (m *Management) GetOrgbyID(ctx echo.Context, orgID int64) error {
	return nil
}

// (PUT /organizations/{orgID})
func (m *Management) UpdateOrganization(ctx echo.Context, orgID int64) error {
	return nil
}
