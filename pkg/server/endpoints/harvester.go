package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	chttp "github.com/HewlettPackard/galadriel/pkg/common/http"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	harvesterapi "github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

const tokenKey = "token"

type HarvesterAPIHandlers struct {
	Logger    logrus.FieldLogger
	Datastore datastore.Datastore
}

// GetRelationships list all the relationships - (GET /relationships)
func (h HarvesterAPIHandlers) GetRelationships(ctx echo.Context, params harvesterapi.GetRelationshipsParams) error {
	return nil
}

// PatchRelationshipsRelationshipID accept/denied relationships requests - (PATCH /relationships/{relationshipID})
func (h HarvesterAPIHandlers) PatchRelationshipsRelationshipID(ctx echo.Context, relationshipID api.UUID) error {
	return nil
}

// Onboard authenticate a trust domain in galadriel server providing its access token - (POST /trust-domain/onboard)
func (h HarvesterAPIHandlers) Onboard(ctx echo.Context) error {
	return nil
}

// BundleSync synchronize the status of trust bundles between server and harvester - (POST /trust-domain/{trustDomainName}/bundles/sync)
func (h HarvesterAPIHandlers) BundleSync(ctx echo.Context, trustDomainName api.TrustDomainName) error {
	return nil
}

// BundlePut uploads a new trust bundle to the server  - (PUT /trust-domain/{trustDomainName}/bundles)
func (h HarvesterAPIHandlers) BundlePut(ctx echo.Context, trustDomainName api.TrustDomainName) error {
	h.Logger.Debug("Receiving post bundle request")
	gctx := ctx.Request().Context()

	// TODO: move authn out and replace with Access Token when implemented
	jt, ok := ctx.Get(tokenKey).(*entity.JoinToken)
	if !ok {
		err := errors.New("error parsing token")
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	token, err := h.Datastore.FindJoinToken(gctx, jt.Token)
	if err != nil {
		err := errors.New("error looking up token")
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	authenticatedTD, err := h.Datastore.FindTrustDomainByID(gctx, token.TrustDomainID)
	if err != nil {
		err := errors.New("error looking up trust domain")
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	if authenticatedTD.Name.String() != trustDomainName {
		return fmt.Errorf("authenticated trust domain {%s} does not match trust domain in path: {%s}", authenticatedTD.Name, trustDomainName)
	}
	// end authn

	req := &harvesterapi.BundlePutJSONRequestBody{}
	err = chttp.FromBody(ctx, req)
	if err != nil {
		err := fmt.Errorf("failed to read bundle put body: %v", err)
		return h.handleErrorAndLog(err, http.StatusBadRequest)
	}

	if authenticatedTD.Name.String() != req.TrustDomain {
		err := fmt.Errorf("authenticated trust domain {%s} does not match trust domain in request body: {%s}", authenticatedTD.Name, req.TrustDomain)
		return h.handleErrorAndLog(err, http.StatusBadRequest)
	}

	storedBundle, err := h.Datastore.FindBundleByTrustDomainID(gctx, authenticatedTD.ID.UUID)
	if err != nil || storedBundle == nil {
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	if req.TrustBundle == "" {
		return nil
	}

	bundle, err := req.ToEntity()
	if err != nil {
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	if storedBundle != nil {
		bundle.TrustDomainID = storedBundle.TrustDomainID
	}

	_, err = h.Datastore.CreateOrUpdateBundle(gctx, bundle)
	if err != nil {
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	if err = chttp.BodylessResponse(ctx); err != nil {
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

func (h HarvesterAPIHandlers) handleErrorAndLog(err error, code int) error {
	errMsg := util.LogSanitize(err.Error())
	h.Logger.Errorf(errMsg)
	return echo.NewHTTPError(code, err.Error())
}

type harvesterAPIHandlers struct {
	logger    logrus.FieldLogger
	datastore datastore.Datastore
}

// NewHarvesterAPIHandlers create a new harvesterAPIHandlers
func NewHarvesterAPIHandlers(l logrus.FieldLogger, ds datastore.Datastore) *harvesterAPIHandlers {
	return &harvesterAPIHandlers{
		logger:    l,
		datastore: ds,
	}
}

func (h harvesterAPIHandlers) handleErrorAndLog(err error, code int) error {
	errMsg := util.LogSanitize(err.Error())
	h.logger.Errorf(errMsg)
	return echo.NewHTTPError(code, err.Error())
}

// GetRelationships lists all the relationships for a given Trust Domain
func (h *harvesterAPIHandlers) GetRelationships(ctx echo.Context, params harvesterapi.GetRelationshipsParams) error {
	if params.TrustDomainName == nil {
		err := errors.New("trust domain name is required")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	tdName, err := spiffeid.TrustDomainFromString(*params.TrustDomainName)
	if err != nil {
		err := errors.New("error parsing trust domain name")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	td, err := h.datastore.FindTrustDomainByName(ctx.Request().Context(), tdName)
	if err != nil {
		err := errors.New("error looking up trust domain")
		h.handleTCPError(ctx, err.Error())
		return err
	}
	if td == nil {
		// TODO: handle all errors properly, return a valid json with 404 code
		err := errors.New("trust domain not found")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	rels, err := h.datastore.FindRelationshipsByTrustDomainID(ctx.Request().Context(), td.ID.UUID)
	if err != nil {
		err := errors.New("error looking up relationships")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	if params.Status != nil {
		rels = filterByStatus(rels, *params.Status)
	}

	relationships := make([]*api.Relationship, 0, len(rels))
	for _, rel := range rels {
		relationships = append(relationships, harvesterapi.RelationshipFromEntity(rel))
	}

	if err := WriteResponse(ctx, relationships); err != nil {
		err := errors.New("error writing response")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	return nil
}

// PatchRelationshipsRelationshipID updates an existing relationship between two trust domains
func (h *harvesterAPIHandlers) PatchRelationshipsRelationshipID(ctx echo.Context, relationshipID uuid.UUID) error {
	if relationshipID == uuid.Nil {
		err := errors.New("relationship id is required")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	existingRel, err := h.datastore.FindRelationshipByID(ctx.Request().Context(), relationshipID)
	if err != nil {
		err := errors.New("error looking up relationship")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	if existingRel == nil {
		// TODO: handle all errors properly, return a valid json with 404 code// TODO: handle all errors properly, return a valid json with 404 code
		err := errors.New("relationship not found")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	relApproval, err := chttp.FromBody2[harvesterapi.RelationshipApproval](ctx)
	if err != nil {
		err := errors.New("error parsing request body")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	jt, ok := ctx.Get(tokenKey).(*entity.JoinToken)
	if !ok {
		err := errors.New("error parsing join token")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	// TODO: should we check the current status? can an already denied relationship be accepted?
	switch jt.TrustDomainID {
	case existingRel.TrustDomainAID:
		existingRel.TrustDomainAConsent = relApproval.Accept
	case existingRel.TrustDomainBID:
		existingRel.TrustDomainBConsent = relApproval.Accept
	}

	var updatedRel *entity.Relationship
	if updatedRel, err = h.datastore.CreateOrUpdateRelationship(ctx.Request().Context(), existingRel); err != nil {
		err := errors.New("error updating relationship")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	updatedApiRel := harvesterapi.RelationshipFromEntity(updatedRel)
	if err := WriteResponse(ctx, updatedApiRel); err != nil {
		err := errors.New("error writing response")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	return nil
}

func (h *harvesterAPIHandlers) Onboard(ctx echo.Context) error {
	return nil
}

// BundleSync returns the updated bundles for a given trust domain driven by its relationships
func (h *harvesterAPIHandlers) BundleSync(ctx echo.Context, trustDomainName api.TrustDomainName) error {
	h.logger.Debug("Receiving sync request")

	// TODO: move authn out and replace with Access Token when implemented
	jt, ok := ctx.Get(tokenKey).(*entity.JoinToken)
	if !ok {
		err := errors.New("error parsing join token")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	token, err := h.datastore.FindJoinToken(ctx.Request().Context(), jt.Token)
	if err != nil {
		err := errors.New("error looking up token")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	harvesterTrustDomain, err := h.datastore.FindTrustDomainByID(ctx.Request().Context(), token.TrustDomainID)
	if err != nil {
		h.handleTCPError(ctx, fmt.Sprintf("failed to read body: %v", err))
		return err
	}
	// end authn

	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		h.handleTCPError(ctx, fmt.Sprintf("failed to read body: %v", err))
		return err
	}

	receivedHarvesterState := common.SyncBundleRequest{}
	err = json.Unmarshal(body, &receivedHarvesterState)
	if err != nil {
		h.handleTCPError(ctx, fmt.Sprintf("failed to unmarshal state: %v", err))
		return err
	}

	harvesterBundleDigests := receivedHarvesterState.State

	_, foundSelf := receivedHarvesterState.State[harvesterTrustDomain.Name]
	if foundSelf {
		h.handleTCPError(ctx, "bad request: harvester cannot federate with itself")
		return err
	}

	relationships, err := h.datastore.FindRelationshipsByTrustDomainID(ctx.Request().Context(), harvesterTrustDomain.ID.UUID)
	if err != nil {
		h.handleTCPError(ctx, fmt.Sprintf("failed to fetch relationships: %v", err))
		return err
	}

	federatedTDs := getFederatedTrustDomains(relationships, harvesterTrustDomain.ID.UUID)

	if len(federatedTDs) == 0 {
		h.logger.Debug("No federated trust domains yet")
		return nil
	}

	federatedBundles, federatedBundlesDigests, err := h.getCurrentFederatedBundles(ctx.Request().Context(), federatedTDs)
	if err != nil {
		h.handleTCPError(ctx, fmt.Sprintf("failed to fetch bundles from DB: %v", err))
		return err
	}

	if len(federatedBundles) == 0 {
		h.logger.Debug("No federated bundles yet")
		return nil
	}

	bundlesUpdates, err := h.getFederatedBundlesUpdates(ctx.Request().Context(), harvesterBundleDigests, federatedBundles)
	if err != nil {
		h.handleTCPError(ctx, fmt.Sprintf("failed to fetch bundles from DB: %v", err))
		return err
	}

	response := common.SyncBundleResponse{
		Updates: bundlesUpdates,
		State:   federatedBundlesDigests,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		h.handleTCPError(ctx, fmt.Sprintf("failed to marshal response: %v", err))
		return err
	}

	_, err = ctx.Response().Write(responseBytes)
	if err != nil {
		h.handleTCPError(ctx, fmt.Sprintf("failed to write response: %v", err))
		return err
	}

	return nil
}

// BundlePut updates the bundle for a given trust domain
func (h *harvesterAPIHandlers) BundlePut(ctx echo.Context, trustDomainName api.TrustDomainName) error {
	h.logger.Debug("Receiving post bundle request")

	// TODO: move authn out and replace with Access Token when implemented
	jt, ok := ctx.Get(tokenKey).(*entity.JoinToken)
	if !ok {
		err := errors.New("error parsing token")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	token, err := h.datastore.FindJoinToken(ctx.Request().Context(), jt.Token)
	if err != nil {
		err := errors.New("error looking up token")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	authenticatedTD, err := h.datastore.FindTrustDomainByID(ctx.Request().Context(), token.TrustDomainID)
	if err != nil {
		err := errors.New("error looking up trust domain")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	if authenticatedTD.Name.String() != trustDomainName {
		return fmt.Errorf("authenticated trust domain {%s} does not match trust domain in path: {%s}", authenticatedTD.Name, trustDomainName)
	}
	// end authn

	var req harvesterapi.BundlePut
	err = chttp.FromBody(ctx, req)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal request body: %v", err)
		return h.handleErrorAndLog(err, http.StatusBadRequest)
	}

	if authenticatedTD.Name.String() != req.TrustDomain {
		err := fmt.Errorf("authenticated trust domain {%s} does not match trust domain in request body: {%s}", authenticatedTD.Name, req.TrustDomain)
		h.handleTCPError(ctx, err.Error())
		return err
	}

	storedBundle, err := h.datastore.FindBundleByTrustDomainID(ctx.Request().Context(), authenticatedTD.ID.UUID)
	if err != nil || storedBundle == nil {
		err := errors.New("error looking up bundle")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	if req.TrustBundle == "" {
		err := errors.New("trust bundle is not set")
		h.handleTCPError(ctx, err.Error())
		return err
	}

	bundle, err := req.ToEntity()
	if err != nil {
		err := errors.New("error parsing request")
		h.handleTCPError(ctx, err.Error())
		return err
	}
	bundle.TrustDomainID = storedBundle.TrustDomainID

	res, err := h.datastore.CreateOrUpdateBundle(ctx.Request().Context(), bundle)
	if err != nil {
		h.handleTCPError(ctx, err.Error())
		return err
	}

	if err = WriteResponse(ctx, res); err != nil {
		h.handleTCPError(ctx, err.Error())
		return err
	}

	return nil
}

func (h *harvesterAPIHandlers) handleTCPError(ctx echo.Context, errMsg string) {
	h.logger.Errorf(errMsg)
	_, err := ctx.Response().Write([]byte(errMsg))
	if err != nil {
		h.logger.Errorf("Failed to write error response: %v", err)
	}
}

func (h *harvesterAPIHandlers) getCurrentFederatedBundles(ctx context.Context, federatedTDs []uuid.UUID) ([]*entity.Bundle, common.BundlesDigests, error) {
	var bundles []*entity.Bundle
	bundlesDigests := map[spiffeid.TrustDomain][]byte{}

	for _, id := range federatedTDs {
		b, err := h.datastore.FindBundleByTrustDomainID(ctx, id)
		if err != nil {
			return nil, nil, err
		}
		td, err := h.datastore.FindTrustDomainByID(ctx, id)
		if err != nil {
			return nil, nil, err
		}

		if b != nil {
			bundles = append(bundles, b)
			bundlesDigests[td.Name] = util.GetDigest(b.Data)
		}
	}

	return bundles, bundlesDigests, nil
}

func (h *harvesterAPIHandlers) getFederatedBundlesUpdates(ctx context.Context, harvesterBundlesDigests common.BundlesDigests, federatedBundles []*entity.Bundle) (common.BundleUpdates, error) {
	response := make(common.BundleUpdates)

	for _, b := range federatedBundles {
		td, err := h.datastore.FindTrustDomainByID(ctx, b.TrustDomainID)
		if err != nil {
			return nil, err
		}

		serverDigest := util.GetDigest(b.Data)
		harvesterDigest := harvesterBundlesDigests[td.Name]

		// If the bundle digest received from a federated trust domain of the calling harvester is not the same as the
		// digest the server has, the harvester needs to be updated of the new bundle. This also covers th case of
		// the harvester not being aware of any bundles. The update represents a newly federated trustDomain's bundle.
		if !bytes.Equal(harvesterDigest, serverDigest) {
			response[td.Name] = b
		}
	}

	return response, nil
}

// WritesReponse writes the response to the client
func WriteResponse(ctx echo.Context, body interface{}) error {
	if body == nil {
		return nil
	}

	bytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal response body: %v", err)
	}
	ctx.Response().Header().Add("Content-Type", "application/json")
	_, err = ctx.Response().Write(bytes)
	if err != nil {
		return fmt.Errorf("failed to write response body: %v", err)
	}

	return nil
}

func getFederatedTrustDomains(relationships []*entity.Relationship, tdID uuid.UUID) []uuid.UUID {
	var federatedTrustDomains []uuid.UUID

	for _, r := range relationships {
		ma := r.TrustDomainAID
		mb := r.TrustDomainBID

		if tdID == ma {
			federatedTrustDomains = append(federatedTrustDomains, mb)
		} else {
			federatedTrustDomains = append(federatedTrustDomains, ma)
		}
	}
	return federatedTrustDomains
}

func filterByStatus(rels []*entity.Relationship, status harvesterapi.GetRelationshipsParamsStatus) []*entity.Relationship {
	var filteredRels []*entity.Relationship
	for _, rel := range rels {
		if getStatus(rel) == status {
			filteredRels = append(filteredRels, rel)
		}
	}

	return filteredRels
}

func getStatus(rel *entity.Relationship) harvesterapi.GetRelationshipsParamsStatus {
	/*
		TODO: review this, we need to have the consents to be nullable
		if we want to support these three statuses

		null: no response given
		false: harvester explicitly denied the relationship
		true: harvester explicitly accepted the relationship
	*/
	switch {
	case rel.TrustDomainAConsent && rel.TrustDomainBConsent:
		return harvesterapi.Accepted
	case rel.TrustDomainAConsent || rel.TrustDomainBConsent:
		return harvesterapi.Pending
	default:
		return harvesterapi.Denied
	}
}
