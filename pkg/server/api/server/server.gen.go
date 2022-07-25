// Package server provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.0 DO NOT EDIT.
package server

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// Defines values for FederationGroupStatus.
const (
	FederationGroupStatusActive   FederationGroupStatus = "active"
	FederationGroupStatusInactive FederationGroupStatus = "inactive"
)

// Defines values for FederationGroupMembershipStatus.
const (
	FederationGroupMembershipStatusActive   FederationGroupMembershipStatus = "active"
	FederationGroupMembershipStatusInactive FederationGroupMembershipStatus = "inactive"
)

// Defines values for SpireServerStatus.
const (
	Active   SpireServerStatus = "active"
	Inactive SpireServerStatus = "inactive"
	Invited  SpireServerStatus = "invited"
)

// FederationGroup defines model for FederationGroup.
type FederationGroup struct {
	Id     int64                  `json:"id"`
	Name   string                 `json:"name"`
	Orgid  int64                  `json:"orgid"`
	Status *FederationGroupStatus `json:"status,omitempty"`
}

// FederationGroupStatus defines model for FederationGroup.Status.
type FederationGroupStatus string

// FederationGroupMembership defines model for FederationGroupMembership.
type FederationGroupMembership struct {
	FederationGroupId int64                           `json:"federationGroupId"`
	Id                *int64                          `json:"id,omitempty"`
	SpireServerId     int64                           `json:"spireServerId"`
	Status            FederationGroupMembershipStatus `json:"status"`
}

// FederationGroupMembershipStatus defines model for FederationGroupMembership.Status.
type FederationGroupMembershipStatus string

// Organization defines model for Organization.
type Organization struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

// SpireServer defines model for SpireServer.
type SpireServer struct {
	Description string            `json:"description"`
	Id          int64             `json:"id"`
	Status      SpireServerStatus `json:"status"`
	TrustDomain *string           `json:"trustDomain,omitempty"`
}

// SpireServerStatus defines model for SpireServer.Status.
type SpireServerStatus string

// GetFederationGroupMembershipsParams defines parameters for GetFederationGroupMemberships.
type GetFederationGroupMembershipsParams struct {
	// filter federation groups memberships by orgId
	OrgId *string `form:"orgId,omitempty" json:"orgId,omitempty"`

	// filter federation groups memberships by orgname
	Orgname *string `form:"orgname,omitempty" json:"orgname,omitempty"`

	// filter memberships by trust domain (SpireSever)
	TrustDomain *string `form:"trustDomain,omitempty" json:"trustDomain,omitempty"`

	// filter memberships by status
	Status *GetFederationGroupMembershipsParamsStatus `form:"status,omitempty" json:"status,omitempty"`
}

// GetFederationGroupMembershipsParamsStatus defines parameters for GetFederationGroupMemberships.
type GetFederationGroupMembershipsParamsStatus string

// GetFederationGroupsParams defines parameters for GetFederationGroups.
type GetFederationGroupsParams struct {
	// filter federation groups by orgId
	OrgId *string `form:"orgId,omitempty" json:"orgId,omitempty"`

	// filter federation groups by orgname
	Orgname *string `form:"orgname,omitempty" json:"orgname,omitempty"`

	// filter organizations by federation group name
	Name *string `form:"name,omitempty" json:"name,omitempty"`
}

// GetFederationRelationshipsParams defines parameters for GetFederationRelationships.
type GetFederationRelationshipsParams struct {
	// filter federation groups memberships by orgId
	OrgId *string `form:"orgId,omitempty" json:"orgId,omitempty"`

	// filter federation groups memberships by orgname
	Orgname *string `form:"orgname,omitempty" json:"orgname,omitempty"`

	// filter memberships by trust domain (SpireSever)
	TrustDomain *string `form:"trustDomain,omitempty" json:"trustDomain,omitempty"`

	// filter memberships by status
	Status *GetFederationRelationshipsParamsStatus `form:"status,omitempty" json:"status,omitempty"`
}

// GetFederationRelationshipsParamsStatus defines parameters for GetFederationRelationships.
type GetFederationRelationshipsParamsStatus string

// GetOrganizationsParams defines parameters for GetOrganizations.
type GetOrganizationsParams struct {
	// filter organizations by name
	Name *string `form:"name,omitempty" json:"name,omitempty"`
}

// GetSpireServersParams defines parameters for GetSpireServers.
type GetSpireServersParams struct {
	// filter SpireServers by trust domain
	TrustDomain *string `form:"trustDomain,omitempty" json:"trustDomain,omitempty"`

	// filter SpireServers by status
	Status *GetSpireServersParamsStatus `form:"status,omitempty" json:"status,omitempty"`
}

// GetSpireServersParamsStatus defines parameters for GetSpireServers.
type GetSpireServersParamsStatus string

// CreateSpireServerJSONBody defines parameters for CreateSpireServer.
type CreateSpireServerJSONBody = SpireServer

// CreateSpireServerJSONRequestBody defines body for CreateSpireServer for application/json ContentType.
type CreateSpireServerJSONRequestBody = CreateSpireServerJSONBody

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /federationGroupMemberships)
	GetFederationGroupMemberships(ctx echo.Context, params GetFederationGroupMembershipsParams) error

	// (POST /federationGroupMemberships)
	CreateFederationGroupMembership(ctx echo.Context) error

	// (DELETE /federationGroupMemberships/{membershipID})
	DeletefederationGroupMembership(ctx echo.Context, membershipID int64) error

	// (GET /federationGroupMemberships/{membershipID})
	GetFederationGroupMembershipbyID(ctx echo.Context, membershipID int64) error

	// (PUT /federationGroupMemberships/{membershipID})
	UpdatefederationGroupMembership(ctx echo.Context, membershipID int64) error

	// (GET /federationGroups)
	GetFederationGroups(ctx echo.Context, params GetFederationGroupsParams) error

	// (POST /federationGroups)
	CreateFederationGroup(ctx echo.Context) error

	// (DELETE /federationGroups/{federationGroupID})
	DeletefederationGroup(ctx echo.Context, federationGroupID int64) error

	// (GET /federationGroups/{federationGroupID})
	GetFederationGroupbyID(ctx echo.Context, federationGroupID int64) error

	// (PUT /federationGroups/{federationGroupID})
	UpdatefederationGroup(ctx echo.Context, federationGroupID int64) error

	// (GET /federationRelationships)
	GetFederationRelationships(ctx echo.Context, params GetFederationRelationshipsParams) error

	// (POST /federationRelationships)
	CreateFederationRelationship(ctx echo.Context) error

	// (GET /federationRelationships/{relationshipID})
	GetFederationRelationshipbyID(ctx echo.Context, relationshipID int64) error

	// (PUT /federationRelationships/{relationshipID})
	UpdateFederationRelationshipship(ctx echo.Context, relationshipID int64) error

	// (GET /organizations)
	GetOrganizations(ctx echo.Context, params GetOrganizationsParams) error

	// (POST /organizations)
	CreateOrganization(ctx echo.Context) error

	// (DELETE /organizations/{orgID})
	DeleteOrganization(ctx echo.Context, orgID int64) error

	// (GET /organizations/{orgID})
	GetOrgbyID(ctx echo.Context, orgID int64) error

	// (PUT /organizations/{orgID})
	UpdateOrganizaion(ctx echo.Context, orgID int64) error

	// (GET /spireServers)
	GetSpireServers(ctx echo.Context, params GetSpireServersParams) error

	// (POST /spireServers)
	CreateSpireServer(ctx echo.Context) error

	// (DELETE /spireServers/{spireServerId})
	DeleteSpireServer(ctx echo.Context, spireServerId int64) error

	// (PUT /spireServers/{spireServerId})
	UpdateSpireServer(ctx echo.Context, spireServerId int64) error

	// (PUT /trustBundles/{trustBundleId})
	UpdateTrustBundle(ctx echo.Context, trustBundleId int64) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetFederationGroupMemberships converts echo context to params.
func (w *ServerInterfaceWrapper) GetFederationGroupMemberships(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetFederationGroupMembershipsParams
	// ------------- Optional query parameter "orgId" -------------

	err = runtime.BindQueryParameter("form", true, false, "orgId", ctx.QueryParams(), &params.OrgId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter orgId: %s", err))
	}

	// ------------- Optional query parameter "orgname" -------------

	err = runtime.BindQueryParameter("form", true, false, "orgname", ctx.QueryParams(), &params.Orgname)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter orgname: %s", err))
	}

	// ------------- Optional query parameter "trustDomain" -------------

	err = runtime.BindQueryParameter("form", true, false, "trustDomain", ctx.QueryParams(), &params.TrustDomain)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter trustDomain: %s", err))
	}

	// ------------- Optional query parameter "status" -------------

	err = runtime.BindQueryParameter("form", true, false, "status", ctx.QueryParams(), &params.Status)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter status: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetFederationGroupMemberships(ctx, params)
	return err
}

// CreateFederationGroupMembership converts echo context to params.
func (w *ServerInterfaceWrapper) CreateFederationGroupMembership(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateFederationGroupMembership(ctx)
	return err
}

// DeletefederationGroupMembership converts echo context to params.
func (w *ServerInterfaceWrapper) DeletefederationGroupMembership(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "membershipID" -------------
	var membershipID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "membershipID", runtime.ParamLocationPath, ctx.Param("membershipID"), &membershipID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter membershipID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DeletefederationGroupMembership(ctx, membershipID)
	return err
}

// GetFederationGroupMembershipbyID converts echo context to params.
func (w *ServerInterfaceWrapper) GetFederationGroupMembershipbyID(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "membershipID" -------------
	var membershipID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "membershipID", runtime.ParamLocationPath, ctx.Param("membershipID"), &membershipID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter membershipID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetFederationGroupMembershipbyID(ctx, membershipID)
	return err
}

// UpdatefederationGroupMembership converts echo context to params.
func (w *ServerInterfaceWrapper) UpdatefederationGroupMembership(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "membershipID" -------------
	var membershipID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "membershipID", runtime.ParamLocationPath, ctx.Param("membershipID"), &membershipID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter membershipID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdatefederationGroupMembership(ctx, membershipID)
	return err
}

// GetFederationGroups converts echo context to params.
func (w *ServerInterfaceWrapper) GetFederationGroups(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetFederationGroupsParams
	// ------------- Optional query parameter "orgId" -------------

	err = runtime.BindQueryParameter("form", true, false, "orgId", ctx.QueryParams(), &params.OrgId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter orgId: %s", err))
	}

	// ------------- Optional query parameter "orgname" -------------

	err = runtime.BindQueryParameter("form", true, false, "orgname", ctx.QueryParams(), &params.Orgname)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter orgname: %s", err))
	}

	// ------------- Optional query parameter "name" -------------

	err = runtime.BindQueryParameter("form", true, false, "name", ctx.QueryParams(), &params.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetFederationGroups(ctx, params)
	return err
}

// CreateFederationGroup converts echo context to params.
func (w *ServerInterfaceWrapper) CreateFederationGroup(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateFederationGroup(ctx)
	return err
}

// DeletefederationGroup converts echo context to params.
func (w *ServerInterfaceWrapper) DeletefederationGroup(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "federationGroupID" -------------
	var federationGroupID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "federationGroupID", runtime.ParamLocationPath, ctx.Param("federationGroupID"), &federationGroupID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter federationGroupID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DeletefederationGroup(ctx, federationGroupID)
	return err
}

// GetFederationGroupbyID converts echo context to params.
func (w *ServerInterfaceWrapper) GetFederationGroupbyID(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "federationGroupID" -------------
	var federationGroupID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "federationGroupID", runtime.ParamLocationPath, ctx.Param("federationGroupID"), &federationGroupID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter federationGroupID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetFederationGroupbyID(ctx, federationGroupID)
	return err
}

// UpdatefederationGroup converts echo context to params.
func (w *ServerInterfaceWrapper) UpdatefederationGroup(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "federationGroupID" -------------
	var federationGroupID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "federationGroupID", runtime.ParamLocationPath, ctx.Param("federationGroupID"), &federationGroupID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter federationGroupID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdatefederationGroup(ctx, federationGroupID)
	return err
}

// GetFederationRelationships converts echo context to params.
func (w *ServerInterfaceWrapper) GetFederationRelationships(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetFederationRelationshipsParams
	// ------------- Optional query parameter "orgId" -------------

	err = runtime.BindQueryParameter("form", true, false, "orgId", ctx.QueryParams(), &params.OrgId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter orgId: %s", err))
	}

	// ------------- Optional query parameter "orgname" -------------

	err = runtime.BindQueryParameter("form", true, false, "orgname", ctx.QueryParams(), &params.Orgname)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter orgname: %s", err))
	}

	// ------------- Optional query parameter "trustDomain" -------------

	err = runtime.BindQueryParameter("form", true, false, "trustDomain", ctx.QueryParams(), &params.TrustDomain)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter trustDomain: %s", err))
	}

	// ------------- Optional query parameter "status" -------------

	err = runtime.BindQueryParameter("form", true, false, "status", ctx.QueryParams(), &params.Status)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter status: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetFederationRelationships(ctx, params)
	return err
}

// CreateFederationRelationship converts echo context to params.
func (w *ServerInterfaceWrapper) CreateFederationRelationship(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateFederationRelationship(ctx)
	return err
}

// GetFederationRelationshipbyID converts echo context to params.
func (w *ServerInterfaceWrapper) GetFederationRelationshipbyID(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "relationshipID" -------------
	var relationshipID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "relationshipID", runtime.ParamLocationPath, ctx.Param("relationshipID"), &relationshipID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter relationshipID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetFederationRelationshipbyID(ctx, relationshipID)
	return err
}

// UpdateFederationRelationshipship converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateFederationRelationshipship(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "relationshipID" -------------
	var relationshipID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "relationshipID", runtime.ParamLocationPath, ctx.Param("relationshipID"), &relationshipID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter relationshipID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateFederationRelationshipship(ctx, relationshipID)
	return err
}

// GetOrganizations converts echo context to params.
func (w *ServerInterfaceWrapper) GetOrganizations(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetOrganizationsParams
	// ------------- Optional query parameter "name" -------------

	err = runtime.BindQueryParameter("form", true, false, "name", ctx.QueryParams(), &params.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetOrganizations(ctx, params)
	return err
}

// CreateOrganization converts echo context to params.
func (w *ServerInterfaceWrapper) CreateOrganization(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateOrganization(ctx)
	return err
}

// DeleteOrganization converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteOrganization(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "orgID" -------------
	var orgID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "orgID", runtime.ParamLocationPath, ctx.Param("orgID"), &orgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter orgID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DeleteOrganization(ctx, orgID)
	return err
}

// GetOrgbyID converts echo context to params.
func (w *ServerInterfaceWrapper) GetOrgbyID(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "orgID" -------------
	var orgID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "orgID", runtime.ParamLocationPath, ctx.Param("orgID"), &orgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter orgID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetOrgbyID(ctx, orgID)
	return err
}

// UpdateOrganizaion converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateOrganizaion(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "orgID" -------------
	var orgID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "orgID", runtime.ParamLocationPath, ctx.Param("orgID"), &orgID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter orgID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateOrganizaion(ctx, orgID)
	return err
}

// GetSpireServers converts echo context to params.
func (w *ServerInterfaceWrapper) GetSpireServers(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetSpireServersParams
	// ------------- Optional query parameter "trustDomain" -------------

	err = runtime.BindQueryParameter("form", true, false, "trustDomain", ctx.QueryParams(), &params.TrustDomain)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter trustDomain: %s", err))
	}

	// ------------- Optional query parameter "status" -------------

	err = runtime.BindQueryParameter("form", true, false, "status", ctx.QueryParams(), &params.Status)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter status: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetSpireServers(ctx, params)
	return err
}

// CreateSpireServer converts echo context to params.
func (w *ServerInterfaceWrapper) CreateSpireServer(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateSpireServer(ctx)
	return err
}

// DeleteSpireServer converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteSpireServer(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "spireServerId" -------------
	var spireServerId int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "spireServerId", runtime.ParamLocationPath, ctx.Param("spireServerId"), &spireServerId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter spireServerId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DeleteSpireServer(ctx, spireServerId)
	return err
}

// UpdateSpireServer converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateSpireServer(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "spireServerId" -------------
	var spireServerId int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "spireServerId", runtime.ParamLocationPath, ctx.Param("spireServerId"), &spireServerId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter spireServerId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateSpireServer(ctx, spireServerId)
	return err
}

// UpdateTrustBundle converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateTrustBundle(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "trustBundleId" -------------
	var trustBundleId int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "trustBundleId", runtime.ParamLocationPath, ctx.Param("trustBundleId"), &trustBundleId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter trustBundleId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateTrustBundle(ctx, trustBundleId)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/federationGroupMemberships", wrapper.GetFederationGroupMemberships)
	router.POST(baseURL+"/federationGroupMemberships", wrapper.CreateFederationGroupMembership)
	router.DELETE(baseURL+"/federationGroupMemberships/:membershipID", wrapper.DeletefederationGroupMembership)
	router.GET(baseURL+"/federationGroupMemberships/:membershipID", wrapper.GetFederationGroupMembershipbyID)
	router.PUT(baseURL+"/federationGroupMemberships/:membershipID", wrapper.UpdatefederationGroupMembership)
	router.GET(baseURL+"/federationGroups", wrapper.GetFederationGroups)
	router.POST(baseURL+"/federationGroups", wrapper.CreateFederationGroup)
	router.DELETE(baseURL+"/federationGroups/:federationGroupID", wrapper.DeletefederationGroup)
	router.GET(baseURL+"/federationGroups/:federationGroupID", wrapper.GetFederationGroupbyID)
	router.PUT(baseURL+"/federationGroups/:federationGroupID", wrapper.UpdatefederationGroup)
	router.GET(baseURL+"/federationRelationships", wrapper.GetFederationRelationships)
	router.POST(baseURL+"/federationRelationships", wrapper.CreateFederationRelationship)
	router.GET(baseURL+"/federationRelationships/:relationshipID", wrapper.GetFederationRelationshipbyID)
	router.PUT(baseURL+"/federationRelationships/:relationshipID", wrapper.UpdateFederationRelationshipship)
	router.GET(baseURL+"/organizations", wrapper.GetOrganizations)
	router.POST(baseURL+"/organizations", wrapper.CreateOrganization)
	router.DELETE(baseURL+"/organizations/:orgID", wrapper.DeleteOrganization)
	router.GET(baseURL+"/organizations/:orgID", wrapper.GetOrgbyID)
	router.PUT(baseURL+"/organizations/:orgID", wrapper.UpdateOrganizaion)
	router.GET(baseURL+"/spireServers", wrapper.GetSpireServers)
	router.POST(baseURL+"/spireServers", wrapper.CreateSpireServer)
	router.DELETE(baseURL+"/spireServers/:spireServerId", wrapper.DeleteSpireServer)
	router.PUT(baseURL+"/spireServers/:spireServerId", wrapper.UpdateSpireServer)
	router.PUT(baseURL+"/trustBundles/:trustBundleId", wrapper.UpdateTrustBundle)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"hUAAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
