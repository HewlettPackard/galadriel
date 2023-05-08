package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/HewlettPackard/galadriel/test/jwttest"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
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
	fakeDB := datastore.NewFakeDB()
	logger := logrus.New()
	jwtIssuer, signer := jwttest.NewJWTIssuer()
	jwtValidator := jwttest.NewJWTValidator(signer, []string{"test"})

	return &HarvesterTestSetup{
		EchoCtx:  e.NewContext(req, rec),
		Recorder: rec,
		Handler:  NewHarvesterAPIHandlers(logger, fakeDB, jwtIssuer, jwtValidator),
	}
}

func SetupTrustDomain(t *testing.T, ds datastore.Datastore) (*entity.TrustDomain, error) {
	td, err := spiffeid.TrustDomainFromString(testTrustDomain)
	assert.NoError(t, err)

	tdEntity := &entity.TrustDomain{
		Name:        td,
		Description: "Fake domain",
	}

	return ds.CreateOrUpdateTrustDomain(context.TODO(), tdEntity)
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
	t.Run("Succesfully register bundles for a trust domain", func(t *testing.T) {
		bundlePut := harvester.BundlePut{
			Signature:          "",
			SigningCertificate: "",
			TrustBundle:        "a new bundle",
			TrustDomain:        testTrustDomain,
		}

		body, err := json.Marshal(bundlePut)
		assert.NoError(t, err)

		harvesterTestSetup := NewHarvesterTestSetup(http.MethodPut, "/trust-domain/:trustDomainName/bundles", string(body))
		echoCtx := harvesterTestSetup.EchoCtx

		// Creating Trust Domain
		td, err := SetupTrustDomain(t, harvesterTestSetup.Handler.Datastore)
		assert.NoError(t, err)

		// Creating Auth token to bypass AuthN layer
		assert.NoError(t, err)
		echoCtx.Set(trustDomainKey, td)

		// Test Main Objective
		err = harvesterTestSetup.Handler.BundlePut(echoCtx, testTrustDomain)
		assert.NoError(t, err)

		recorder := harvesterTestSetup.Recorder
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Empty(t, recorder.Body)
	})
}
