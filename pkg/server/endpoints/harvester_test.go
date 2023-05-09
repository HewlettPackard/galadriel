package endpoints

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	tdA           = &entity.TrustDomain{Name: spiffeid.RequireTrustDomainFromString("td-a.org"), ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	tdB           = &entity.TrustDomain{Name: spiffeid.RequireTrustDomainFromString("td-b.org"), ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	tdC           = &entity.TrustDomain{Name: spiffeid.RequireTrustDomainFromString("td-c.org"), ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	pendingRelAB  = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdB.ID.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	pendingRelAC  = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdC.ID.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	approvedRelAB = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdB.ID.UUID, TrustDomainAConsent: true, TrustDomainBConsent: true, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	approvedRelAC = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdC.ID.UUID, TrustDomainAConsent: true, TrustDomainBConsent: true, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	approvedRelBC = &entity.Relationship{TrustDomainAID: tdB.ID.UUID, TrustDomainBID: tdC.ID.UUID, TrustDomainAConsent: true, TrustDomainBConsent: true, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
)

type HarvesterTestSetup struct {
	EchoCtx   echo.Context
	Recorder  *httptest.ResponseRecorder
	Handler   *harvesterAPIHandlers
	Datastore *datastore.FakeDatabase
}

func NewHarvesterTestSetup(method, url, body string) *HarvesterTestSetup {
	e := echo.New()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	fakeDB := datastore.NewFakeDB()
	logger := logrus.New()

	return &HarvesterTestSetup{
		EchoCtx:   e.NewContext(req, rec),
		Recorder:  rec,
		Handler:   NewHarvesterAPIHandlers(logger, fakeDB),
		Datastore: fakeDB,
	}
}

func TestPatchRelationshipRelationshipID(t *testing.T) {
	t.Run("approve relationship", func(t *testing.T) {
		reqParams := &harvester.RelationshipApproval{Accept: true}
		reqBytes, err := json.Marshal(reqParams)
		require.NoError(t, err)

		url := "/relationships/" + tdA.ID.UUID.String()
		setup := NewHarvesterTestSetup(http.MethodPut, url, string(reqBytes))
		setup.Datastore.WithTrustDomains(tdA, tdB)
		setup.Datastore.WithRelationships(pendingRelAB)
		setupToken(t, setup.EchoCtx, setup.Handler.datastore, tdA)

		err = setup.Handler.PatchRelationshipsRelationshipID(setup.EchoCtx, pendingRelAB.ID.UUID)
		assert.NoError(t, err)

		rels, err := setup.Datastore.FindRelationshipsByTrustDomainID(context.Background(), tdA.ID.UUID)
		require.NoError(t, err)
		require.Equal(t, rels[0].TrustDomainAID, tdA.ID.UUID)

		assert.Len(t, rels, 1)
		assert.True(t, rels[0].TrustDomainAConsent)
		assert.False(t, rels[0].TrustDomainBConsent)
	})
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
			TrustBundle: "fake bundle", //TODO: validate bundle
			TrustDomain: tdA.Name.String(),
		}
		body, err := json.Marshal(bundlePut)
		require.NoError(t, err)

		setup := NewHarvesterTestSetup(http.MethodPut, "/trust-domain/:trustDomainName/bundles", string(body))

		setup.Datastore.WithTrustDomains(tdA)
		setupToken(t, setup.EchoCtx, setup.Handler.datastore, tdA)

		err = setup.Handler.BundlePut(setup.EchoCtx, tdA.Name.String())

		assert.NoError(t, err)
		assert.Equal(t, setup.Recorder.Code, http.StatusOK)
		assert.Empty(t, setup.Recorder.Body)
	})
}

func TestGetRelationships(t *testing.T) {
	t.Run("get all relationships", func(t *testing.T) {
		setup := NewHarvesterTestSetup(http.MethodGet, "/relationships", "")

		setup.Datastore.WithTrustDomains(tdA, tdB, tdC)
		setup.Datastore.WithRelationships(approvedRelAB, approvedRelAC, approvedRelBC)
		setupToken(t, setup.EchoCtx, setup.Handler.datastore, tdA)

		tdName := tdA.Name.String()
		params := harvester.GetRelationshipsParams{TrustDomainName: &tdName}

		err := setup.Handler.GetRelationships(setup.EchoCtx, params)
		assert.NoError(t, err)

		bytes, err := io.ReadAll(setup.Recorder.Result().Body)
		require.NoError(t, err)

		var relationships *harvester.RelationshipGet
		err = json.Unmarshal(bytes, &relationships)

		assert.NoError(t, err)
		assert.Equal(t, setup.Recorder.Code, http.StatusOK)
		assert.Len(t, *relationships, 2)
		assert.Condition(t, allRelationshipsBelongToTrustTrusDomain(relationships, tdA.ID.UUID))
	})

	t.Run("get approved relationships", func(t *testing.T) {
		setup := NewHarvesterTestSetup(http.MethodGet, "/relationships", "")

		setup.Datastore.WithTrustDomains(tdA, tdB, tdC)
		setup.Datastore.WithRelationships(approvedRelAB, pendingRelAC)
		setupToken(t, setup.EchoCtx, setup.Handler.datastore, tdA)

		tdName := tdA.Name.String()
		status := harvester.Accepted
		params := harvester.GetRelationshipsParams{
			TrustDomainName: &tdName,
			Status:          &status,
		}

		err := setup.Handler.GetRelationships(setup.EchoCtx, params)
		assert.NoError(t, err)

		bytes, err := io.ReadAll(setup.Recorder.Result().Body)
		require.NoError(t, err)

		var relationships *harvester.RelationshipGet
		err = json.Unmarshal(bytes, &relationships)

		assert.NoError(t, err)
		assert.Equal(t, setup.Recorder.Code, http.StatusOK)
		assert.Len(t, *relationships, 1)
		assert.Condition(t, allRelationshipsBelongToTrustTrusDomain(relationships, tdA.ID.UUID))
	})
}

func allRelationshipsBelongToTrustTrusDomain(relationships *harvester.RelationshipGet, tdID uuid.UUID) func() bool {
	return func() bool {
		for _, rel := range *relationships {
			if rel.TrustDomainAId != tdID && rel.TrustDomainBId != tdID {
				return false
			}
		}
		return true
	}
}
