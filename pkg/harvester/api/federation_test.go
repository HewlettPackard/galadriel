package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/harvester/controller"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var (
	mockFR = &common.FederationRelationship{
		Id: 1,
	}
	mockFRJSON = `
		{"id": 1,"federationGroupId": 1, "spireServerId": 1, "spireServerIdFederatedWith": 2,
		"spireServerConsent": False, "spireServerFederatedWithConsent": True, "status": "active"}
	`
)

func TestGetFederationRelationships(t *testing.T) {
	var controller controller.HarvesterController
	var params harvester.GetFederationRelationshipsParams

	a := NewHTTPApi(controller)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/federation-relationships", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, a.GetFederationRelationships(c, params))
}

func TestGetFederationRelationshipbyId(t *testing.T) {
	var controller controller.HarvesterController
	mockedId := mockFR.Id

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/federation-relationships/%d", mockedId), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	a := NewHTTPApi(controller)
	assert.Error(t, a.GetFederationRelationshipbyId(c, int64(mockedId)))
}

func TestUpdateFederatedRelationshipConsent(t *testing.T) {
	var controller controller.HarvesterController
	mockedId := mockFR.Id

	e := echo.New()

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/federation-relationships/%d", mockedId), strings.NewReader(mockFRJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	a := NewHTTPApi(controller)
	assert.NoError(t, a.UpdateFederatedRelationshipConsent(c, int64(mockedId)))
}
