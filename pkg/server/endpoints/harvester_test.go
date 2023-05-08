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
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type HarvesterTestSetup struct {
	EchoCtx  echo.Context
	Recorder *httptest.ResponseRecorder
	Handler  *harvesterAPIHandlers
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

func setupTrustDomain(t *testing.T, ds datastore.Datastore, trustDomain string) *entity.TrustDomain {
	td, err := spiffeid.TrustDomainFromString(trustDomain)
	require.NoError(t, err)

	tdEntity := &entity.TrustDomain{
		Name:        td,
		Description: "Fake trust domain",
	}
	newTd, err := ds.CreateOrUpdateTrustDomain(context.TODO(), tdEntity)
	require.NoError(t, err)

	return newTd
}

func setupRelationship(t *testing.T, ds datastore.Datastore, tdA *entity.TrustDomain, consentA bool, tdB *entity.TrustDomain, consentB bool) *entity.Relationship {
	rel := &entity.Relationship{
		TrustDomainAID:      tdA.ID.UUID,
		TrustDomainBID:      tdB.ID.UUID,
		TrustDomainAName:    tdA.Name,
		TrustDomainBName:    tdB.Name,
		TrustDomainAConsent: consentA,
		TrustDomainBConsent: consentB,
	}
	rel, err := ds.CreateOrUpdateRelationship(context.TODO(), rel)
	require.NoError(t, err)

	return rel
}

func TestPatchRelationshipRelationshipID(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestOnboard(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestBundleSync(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestBundlePut(t *testing.T) {
	t.Run("register bundles for a trust domain", func(t *testing.T) {
		bundlePut := harvester.BundlePut{
			TrustBundle: "fake bundle",
			TrustDomain: "example.org",
		}

		body, err := json.Marshal(bundlePut)
		assert.NoError(t, err)

		setup := NewHarvesterTestSetup(http.MethodPut, "/trust-domain/:trustDomainName/bundles", string(body))

		// Creating Trust Domain
		td := setupTrustDomain(t, setup.Handler.datastore, "example.org")
		setupToken(t, setup.EchoCtx, setup.Handler.datastore, td)

		// Test Main Objective
		err = setup.Handler.BundlePut(setup.EchoCtx, td.Name.String())

		assert.NoError(t, err)
		assert.Equal(t, setup.Recorder.Code, http.StatusOK)
		assert.Empty(t, setup.Recorder.Body)
	})
}

func TestGetRelationships(t *testing.T) {
	setup := NewHarvesterTestSetup(http.MethodGet, "/relationships", "")

	tdA := setupTrustDomain(t, setup.Handler.datastore, "one.org")
	tdB := setupTrustDomain(t, setup.Handler.datastore, "two.org")
	setupToken(t, setup.EchoCtx, setup.Handler.datastore, tdA)
	setupRelationship(t, setup.Handler.datastore, tdA, true, tdB, true)

	tdName := tdA.Name.String()
	params := harvester.GetRelationshipsParams{
		TrustDomainName: &tdName,
	}

	err := setup.Handler.GetRelationships(setup.EchoCtx, params)
	// assert body content

	assert.NoError(t, err)

}
