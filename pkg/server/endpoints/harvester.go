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

type harvesterAPIHandlers struct {
	logger    logrus.FieldLogger
	datastore datastore.Datastore
}

// NewHarvesterAPIHandlers creates a new harvesterAPIHandlers
func NewHarvesterAPIHandlers(l logrus.FieldLogger, ds datastore.Datastore) *harvesterAPIHandlers {
	return &harvesterAPIHandlers{
		logger:    l,
		datastore: ds,
	}
}

// GetRelationships lists all the relationships for a given Trust Domain
func (h *harvesterAPIHandlers) GetRelationships(ctx echo.Context, params harvesterapi.GetRelationshipsParams) error {
	if params.TrustDomainName == nil {
		err := errors.New("trust domain name is required")
		return h.handleErrorAndLog(err, http.StatusBadRequest)
	}

	tdName, err := spiffeid.TrustDomainFromString(*params.TrustDomainName)
	if err != nil {
		err := errors.New("error parsing trust domain name")
		return h.handleErrorAndLog(err, http.StatusBadRequest)
	}

	td, err := h.datastore.FindTrustDomainByName(ctx.Request().Context(), tdName)
	if err != nil {
		err := errors.New("error looking up trust domain")
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}
	if td == nil {
		// TODO: handle all errors properly, return a valid json with 404 code
		err := errors.New("trust domain not found")
		return h.handleErrorAndLog(err, http.StatusNotFound)
	}

	rels, err := h.datastore.FindRelationshipsByTrustDomainID(ctx.Request().Context(), td.ID.UUID)
	if err != nil {
		err := errors.New("error looking up relationships")
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	if params.Status != nil {
		rels = filterByStatus(rels, *params.Status)
	}

	relationships := make([]*api.Relationship, 0, len(rels))
	for _, rel := range rels {
		relationships = append(relationships, harvesterapi.RelationshipFromEntity(rel))
	}

	if err := chttp.WriteResponse(ctx, relationships); err != nil {
		err := errors.New("error writing response")
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

// PatchRelationshipsRelationshipID updates an existing relationship between two trust domains
func (h *harvesterAPIHandlers) PatchRelationshipsRelationshipID(ctx echo.Context, relationshipID uuid.UUID) error {
	if relationshipID == uuid.Nil {
		err := errors.New("relationship id is required")
		return h.handleErrorAndLog(err, http.StatusBadRequest)
	}

	existingRel, err := h.datastore.FindRelationshipByID(ctx.Request().Context(), relationshipID)
	if err != nil {
		err := errors.New("error looking up relationship")
		return h.handleErrorAndLog(err, http.StatusNotFound)
	}

	if existingRel == nil {
		err := errors.New("relationship not found")
		return h.handleErrorAndLog(err, http.StatusNotFound)
	}

	var relApproval *harvesterapi.RelationshipApproval
	err = chttp.FromBody(ctx, &relApproval)
	if err != nil {
		return h.handleErrorAndLog(err, http.StatusBadRequest)
	}

	jt, ok := ctx.Get(tokenKey).(*entity.JoinToken)
	if !ok {
		err := errors.New("error parsing join token")
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	// TODO: should we check the current status? can an already denied relationship be accepted?
	switch jt.TrustDomainID {
	case existingRel.TrustDomainAID:
		existingRel.TrustDomainAConsent = relApproval.Accept
	case existingRel.TrustDomainBID:
		existingRel.TrustDomainBConsent = relApproval.Accept
	}

	updatedRel, err := h.datastore.CreateOrUpdateRelationship(ctx.Request().Context(), existingRel)
	if err != nil {
		err := errors.New("error updating relationship")
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	updatedApiRel := harvesterapi.RelationshipFromEntity(updatedRel)
	err = chttp.WriteResponse(ctx, updatedApiRel)
	if err != nil {
		err := errors.New("error writing response")
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
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
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	token, err := h.datastore.FindJoinToken(ctx.Request().Context(), jt.Token)
	if err != nil {
		err := errors.New("error looking up token")
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	harvesterTrustDomain, err := h.datastore.FindTrustDomainByID(ctx.Request().Context(), token.TrustDomainID)
	if err != nil {
		err := fmt.Errorf("failed to read body: %v", err)
		return h.handleErrorAndLog(err, http.StatusBadRequest)
	}
	// end authn

	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		err := fmt.Errorf("failed to read body: %v", err)
		return h.handleErrorAndLog(err, http.StatusBadRequest)
	}

	receivedHarvesterState := common.SyncBundleRequest{}
	err = json.Unmarshal(body, &receivedHarvesterState)
	if err != nil {
		err := fmt.Errorf("failed to unmarshal state: %v", err)
		return h.handleErrorAndLog(err, http.StatusBadRequest)
	}

	harvesterBundleDigests := receivedHarvesterState.State

	_, foundSelf := receivedHarvesterState.State[harvesterTrustDomain.Name]
	if foundSelf {
		err := errors.New("bad request: harvester cannot federate with itself")
		return h.handleErrorAndLog(err, http.StatusBadRequest)
	}

	relationships, err := h.datastore.FindRelationshipsByTrustDomainID(ctx.Request().Context(), harvesterTrustDomain.ID.UUID)
	if err != nil {
		err := fmt.Errorf("failed to fetch relationships: %v", err)
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	federatedTDs := getFederatedTrustDomains(relationships, harvesterTrustDomain.ID.UUID)

	if len(federatedTDs) == 0 {
		h.logger.Debug("No federated trust domains yet")
		return nil
	}

	federatedBundles, federatedBundlesDigests, err := h.getCurrentFederatedBundles(ctx.Request().Context(), federatedTDs)
	if err != nil {
		err := fmt.Errorf("failed to fetch bundles: %v", err)
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	if len(federatedBundles) == 0 {
		h.logger.Debug("No federated bundles yet")
		return nil
	}

	bundlesUpdates, err := h.getFederatedBundlesUpdates(ctx.Request().Context(), harvesterBundleDigests, federatedBundles)
	if err != nil {
		err := fmt.Errorf("failed to fetch bundles: %v", err)
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	response := common.SyncBundleResponse{
		Updates: bundlesUpdates,
		State:   federatedBundlesDigests,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		err := fmt.Errorf("failed to marshal response: %v", err)
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	err = chttp.WriteResponse(ctx, responseBytes)
	if err != nil {
		err := fmt.Errorf("failed to write response: %v", err)
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
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
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	token, err := h.datastore.FindJoinToken(ctx.Request().Context(), jt.Token)
	if err != nil {
		err := errors.New("error looking up token")
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	authenticatedTD, err := h.datastore.FindTrustDomainByID(ctx.Request().Context(), token.TrustDomainID)
	if err != nil {
		err := errors.New("error looking up trust domain")
		return h.handleErrorAndLog(err, http.StatusNotFound)
	}

	if authenticatedTD.Name.String() != trustDomainName {
		err := fmt.Errorf("authenticated trust domain {%s} does not match trust domain in path: {%s}", authenticatedTD.Name, trustDomainName)
		return h.handleErrorAndLog(err, http.StatusUnauthorized)
	}
	// end authn

	var req *harvesterapi.BundlePut
	err = chttp.FromBody(ctx, &req)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal request body: %v", err)
		return h.handleErrorAndLog(err, http.StatusBadRequest)
	}

	if authenticatedTD.Name.String() != req.TrustDomain {
		err := fmt.Errorf("authenticated trust domain {%s} does not match trust domain in request body: {%s}", authenticatedTD.Name, req.TrustDomain)
		return h.handleErrorAndLog(err, http.StatusUnauthorized)
	}

	storedBundle, err := h.datastore.FindBundleByTrustDomainID(ctx.Request().Context(), authenticatedTD.ID.UUID)
	if err != nil {
		err := errors.New("error looking up bundle")
		return h.handleErrorAndLog(err, http.StatusNotFound)
	}

	// TODO: validate trust domain format
	if req.TrustBundle == "" {
		err := errors.New("trust bundle is not set")
		return h.handleErrorAndLog(err, http.StatusBadRequest)
	}

	bundle, err := req.ToEntity()
	if err != nil {
		err := errors.New("error parsing request")
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	if storedBundle != nil {
		bundle.TrustDomainID = storedBundle.TrustDomainID
	}

	_, err = h.datastore.CreateOrUpdateBundle(ctx.Request().Context(), bundle)
	if err != nil {
		err := fmt.Errorf("error creating or updating bundle: %v", err)
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	if err = chttp.BodylessResponse(ctx); err != nil {
		return h.handleErrorAndLog(err, http.StatusInternalServerError)
	}

	return nil
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

func (h harvesterAPIHandlers) handleErrorAndLog(err error, code int) error {
	// TODO: return json
	errMsg := util.LogSanitize(err.Error())
	h.logger.Errorf(errMsg)
	return echo.NewHTTPError(code, err.Error())
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
