package endpoints

import (
	"fmt"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	chttp "github.com/HewlettPackard/galadriel/pkg/common/http"
	"github.com/HewlettPackard/galadriel/pkg/harvester/api/admin"
	"github.com/HewlettPackard/galadriel/pkg/harvester/galadrielclient"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type AdminAPIHandlers struct {
	client galadrielclient.Client
	logger logrus.FieldLogger
}

func NewAdminAPIHandlers(logger logrus.FieldLogger, client galadrielclient.Client) *AdminAPIHandlers {
	return &AdminAPIHandlers{
		client: client,
		logger: logger,
	}
}

func (h AdminAPIHandlers) GetRelationships(echoCtx echo.Context, params admin.GetRelationshipsParams) error {
	ctx := echoCtx.Request().Context()

	status := entity.ConsentStatus(*params.ConsentStatus)
	resp, err := h.client.GetRelationships(ctx, status)
	if err != nil {
		return chttp.LogAndRespondWithError(h.logger, err, err.Error(), http.StatusInternalServerError)
	}

	relationships := api.MapRelationships(resp...)
	err = chttp.WriteResponse(echoCtx, http.StatusOK, relationships)
	if err != nil {
		return chttp.LogAndRespondWithError(h.logger, err, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (h AdminAPIHandlers) PatchRelationship(echoCtx echo.Context, relationshipID api.UUID) error {
	ctx := echoCtx.Request().Context()

	reqBody := &admin.PatchRelationshipRequest{}
	err := chttp.ParseRequestBodyToStruct(echoCtx, reqBody)
	if err != nil {
		msg := "failed to read relationship patch body"
		err = fmt.Errorf("%s: %v", msg, err)
		return chttp.LogAndRespondWithError(h.logger, err, err.Error(), http.StatusBadRequest)
	}

	status := entity.ConsentStatus(reqBody.ConsentStatus)
	r, err := h.client.UpdateRelationship(ctx, relationshipID, status)
	if err != nil {
		return chttp.LogAndRespondWithError(h.logger, err, err.Error(), http.StatusInternalServerError)
	}

	rel := api.RelationshipFromEntity(r)
	err = chttp.WriteResponse(echoCtx, http.StatusOK, rel)
	if err != nil {
		return chttp.LogAndRespondWithError(h.logger, err, err.Error(), http.StatusInternalServerError)
	}

	return nil
}
