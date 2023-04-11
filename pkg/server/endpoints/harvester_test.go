package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
)

type HarvesterTestSetup struct {
	EchoCtx  echo.Context
	Handler  *harvesterAPIHandlers
	Recorder *httptest.ResponseRecorder
}

func NewHarvesterTestSetup(method, url, body string) *HarvesterTestSetup {
	e := echo.New()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	fakeDB := datastore.NewFakeDB()
	logger := logrus.New()

	return &HarvesterTestSetup{
		EchoCtx:  e.NewContext(req, rec),
		Recorder: rec,
		Handler:  NewHarvesterAPIHandlers(logger, fakeDB),
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
		td, err := SetupTrustDomain(t, harvesterTestSetup.Handler.datastore)
		assert.NoError(t, err)

		// Creating Auth token to bypass AuthN layer
		token := GenerateSecureToken(10)
		jt := SetupToken(t, harvesterTestSetup.Handler.datastore, token, td.ID.UUID)
		assert.NoError(t, err)
		echoCtx.Set(tokenKey, jt)

		// Test Main Objective
		err = harvesterTestSetup.Handler.BundlePut(echoCtx, testTrustDomain)
		assert.NoError(t, err)

		recorder := harvesterTestSetup.Recorder
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Empty(t, recorder.Body)
	})
}

func TestGetRelationships(t *testing.T) {
	relApproval := &harvester.RelationshipApproval{
		Accept: true,
	}
	body, err := json.Marshal(relApproval)
	assert.NoError(t, err)

	reader := strings.NewReader(string(body))

	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/relationships", reader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ds := datastore.NewFakeDB()
	h := NewHarvesterAPIHandlers(logrus.New(), ds)

	var status harvester.GetRelationshipsParamsStatus = "approved"
	var tdName api.TrustDomainName = "one.org"
	p := harvester.GetRelationshipsParams{
		TrustDomainName: &tdName,
		Status:          &status,
	}

	tdA := &entity.TrustDomain{
		Name:      spiffeid.RequireTrustDomainFromString("one.org"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	tdB := &entity.TrustDomain{
		Name:      spiffeid.RequireTrustDomainFromString("two.org"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = ds.CreateOrUpdateTrustDomain(c.Request().Context(), tdA)
	assert.NoError(t, err)
	_, err = ds.CreateOrUpdateTrustDomain(c.Request().Context(), tdB)
	assert.NoError(t, err)

	rel := &entity.Relationship{
		TrustDomainAID:      tdA.ID.UUID,
		TrustDomainBID:      tdB.ID.UUID,
		TrustDomainAConsent: true,
		TrustDomainBConsent: true,
	}
	_, err = ds.CreateOrUpdateRelationship(c.Request().Context(), rel)
	assert.NoError(t, err)

	ds.CreateJoinToken(c.Request().Context(), &entity.JoinToken{
		TrustDomainID:   tdA.ID.UUID,
		TrustDomainName: tdA.Name,
		Token:           "fake-token2",
	})

	req.Header.Set(echo.HeaderAuthorization, "Bearer fake-token33")

	err = h.GetRelationships(c, p)
	assert.NoError(t, err)

	fmt.Println(rec.Body.String())
}
