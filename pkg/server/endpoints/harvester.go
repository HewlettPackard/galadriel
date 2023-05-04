package endpoints

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	chttp "github.com/HewlettPackard/galadriel/pkg/common/http"
	"github.com/HewlettPackard/galadriel/pkg/common/util"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
)

const tokenKey = "token"

type HarvesterAPIHandlers struct {
	Logger    logrus.FieldLogger
	Datastore datastore.Datastore
}

// NewHarvesterAPIHandlers create a new HarvesterAPIHandlers
func NewHarvesterAPIHandlers(l logrus.FieldLogger, ds datastore.Datastore) *HarvesterAPIHandlers {
	return &HarvesterAPIHandlers{
		Logger:    l,
		Datastore: ds,
	}
}

// GetRelationships list all the relationships - (GET /relationships)
func (h HarvesterAPIHandlers) GetRelationships(ctx echo.Context, params harvester.GetRelationshipsParams) error {
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

	req := &harvester.BundlePutJSONRequestBody{}
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
	if err != nil {
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
