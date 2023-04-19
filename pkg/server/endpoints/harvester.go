package endpoints

import (
	"errors"
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/common/http"
	chttp "github.com/HewlettPackard/galadriel/pkg/common/http"
	"github.com/HewlettPackard/galadriel/pkg/common/util"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	commonAPI "github.com/HewlettPackard/galadriel/pkg/common/api"
	harvesterAPI "github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
)

const tokenKey = "token"

type HarvesterAPIHandlers struct {
	Logger    logrus.FieldLogger
	Datastore datastore.Datastore
}

func NewHarvesterAPIHandlers(l logrus.FieldLogger, ds datastore.Datastore) HarvesterAPIHandlers {
	return HarvesterAPIHandlers{
		Logger:    l,
		Datastore: ds,
	}
}

// List the relationships.
// (GET /relationships)
func (h *HarvesterAPIHandlers) GetRelationships(ctx echo.Context, params harvesterAPI.GetRelationshipsParams) error {
	return nil
}

// Accept/Denies relationship requests
// (PATCH /relationships/{relationshipID})
func (h *HarvesterAPIHandlers) PatchRelationshipsRelationshipID(ctx echo.Context, relationshipID commonAPI.UUID) error {
	return nil
}

// Onboarding a new Trust Domain in the Galadriel Server
// (POST /trust-domain/onboard)
func (h *HarvesterAPIHandlers) Onboard(ctx echo.Context) error {
	return nil
}

// Upload a new trust bundle to the server
// (POST /trust-domain/{trustDomainName}/bundles/sync)
func (h HarvesterAPIHandlers) BundleSync(ctx echo.Context, trustDomainName commonAPI.TrustDomainName) error {
	return nil
}

// Upload a new trust bundle to the server
// (PUT /trust-domain/{trustDomainName}/bundles)
func (h *HarvesterAPIHandlers) BundlePut(ctx echo.Context, trustDomainName commonAPI.TrustDomainName) error {
	h.Logger.Debug("Receiving post bundle request")

	// TODO: move authn out and replace with Access Token when implemented
	jt, ok := ctx.Get(tokenKey).(*entity.JoinToken)
	if !ok {
		err := errors.New("error parsing token")
		h.handleTCPError(ctx, err)
		return err
	}

	token, err := h.Datastore.FindJoinToken(ctx.Request().Context(), jt.Token)
	if err != nil {
		err := errors.New("error looking up token")
		h.handleTCPError(ctx, err)
		return err
	}

	authenticatedTD, err := h.Datastore.FindTrustDomainByID(ctx.Request().Context(), token.TrustDomainID)
	if err != nil {
		err := errors.New("error looking up trust domain")
		h.handleTCPError(ctx, err)
		return err
	}

	if authenticatedTD.Name.String() != trustDomainName {
		return fmt.Errorf("authenticated trust domain {%s} does not match trust domain in path: {%s}", authenticatedTD.Name, trustDomainName)
	}
	// end authn

	req := &harvesterAPI.BundlePut{}

	err = http.FromBody(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to read bundle put body: %v", err)
	}

	if authenticatedTD.Name.String() != req.TrustDomain {
		err := fmt.Errorf("authenticated trust domain {%s} does not match trust domain in request body: {%s}", authenticatedTD.Name, req.TrustDomain)
		h.handleTCPError(ctx, err)
		return err
	}

	storedBundle, err := h.Datastore.FindBundleByTrustDomainID(ctx.Request().Context(), authenticatedTD.ID.UUID)
	if err != nil {
		h.handleTCPError(ctx, err)
		return err
	}

	if req.TrustBundle == "" {
		return nil
	}

	bundle := req.ToEntity()
	if storedBundle != nil {
		bundle.TrustDomainID = storedBundle.TrustDomainID
	}
	res, err := h.Datastore.CreateOrUpdateBundle(ctx.Request().Context(), bundle)
	if err != nil {
		h.handleTCPError(ctx, err)
		return err
	}

	if err = chttp.WriteResponse(ctx, res); err != nil {
		h.handleTCPError(ctx, err)
		return err
	}

	return nil
}

func (h *HarvesterAPIHandlers) handleTCPError(ctx echo.Context, err error) {
	errMsg := util.LogSanitize(err.Error())
	h.Logger.Errorf(errMsg)
	if err := chttp.HandleTCPError(ctx, err); err != nil {
		h.Logger.Errorf(err.Error())
	}
}
