package endpoints

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type HarvesterTestSetup struct {
	EchoCtx  echo.Context
	Handler  *HarvesterAPIHandlers
	Recorder *httptest.ResponseRecorder
}

func NewHarvesterTestSetup(method, url, body string) *HarvesterTestSetup {
	e := echo.New()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	inMemoryDB := datastore.NewInMemoryDB()
	logger := logrus.New()

	return &HarvesterTestSetup{
		EchoCtx:  e.NewContext(req, rec),
		Recorder: rec,
		Handler:  NewHarvesterAPIHandlers(logger, inMemoryDB),
	}
}

func TestTCPGetRelationships(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestTCPPatchRelationshipRelationshipID(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestTCPOnboard(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestTCPBundleSync(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestTCPBundlePut(t *testing.T) {
	t.Error("Need to be implemented")
}
