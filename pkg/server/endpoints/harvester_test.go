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
	"github.com/HewlettPackard/galadriel/test/fakes/fakejwtissuer"
	"github.com/HewlettPackard/galadriel/test/jwttest"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type HarvesterTestSetup struct {
	EchoCtx   echo.Context
	Handler   *HarvesterAPIHandlers
	JWTIssuer *fakejwtissuer.JWTIssuer
	Recorder  *httptest.ResponseRecorder
}

func NewHarvesterTestSetup(t *testing.T, method, url, body string) *HarvesterTestSetup {
	e := echo.New()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	fakeDB := datastore.NewFakeDB()
	logger := logrus.New()

	jwtAudience := []string{"test"}
	jwtIssuer := fakejwtissuer.New(t, "test", testTrustDomain, jwtAudience)
	jwtValidator := jwttest.NewJWTValidator(jwtIssuer.Signer, jwtAudience)

	return &HarvesterTestSetup{
		EchoCtx:   e.NewContext(req, rec),
		Recorder:  rec,
		Handler:   NewHarvesterAPIHandlers(logger, fakeDB, jwtIssuer, jwtValidator),
		JWTIssuer: jwtIssuer,
	}
}

func SetupTrustDomain(t *testing.T, ds datastore.Datastore) *entity.TrustDomain {
	td, err := spiffeid.TrustDomainFromString(testTrustDomain)
	assert.NoError(t, err)

	tdEntity := &entity.TrustDomain{
		Name:        td,
		Description: "Fake domain",
	}

	trustDomain, err := ds.CreateOrUpdateTrustDomain(context.TODO(), tdEntity)
	require.NoError(t, err)

	return trustDomain
}

func SetupJoinToken(t *testing.T, ds datastore.Datastore, td uuid.UUID) *entity.JoinToken {
	jt := &entity.JoinToken{
		Token:         "test-join-token",
		TrustDomainID: td,
	}

	joinToken, err := ds.CreateJoinToken(context.TODO(), jt)
	require.NoError(t, err)

	return joinToken
}

func TestTCPGetRelationships(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestTCPPatchRelationshipRelationshipID(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestTCPOnboard(t *testing.T) {
	t.Run("Successfully onboard a new agent", func(t *testing.T) {
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, "/onboard", "")
		echoCtx := harvesterTestSetup.EchoCtx

		td := SetupTrustDomain(t, harvesterTestSetup.Handler.Datastore)
		token := SetupJoinToken(t, harvesterTestSetup.Handler.Datastore, td.ID.UUID)

		params := harvester.OnboardParams{
			JoinToken: token.Token,
		}
		err := harvesterTestSetup.Handler.Onboard(echoCtx, params)
		assert.NoError(t, err)

		recorder := harvesterTestSetup.Recorder
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.NotEmpty(t, recorder.Body)

		jwtToken := strings.ReplaceAll(recorder.Body.String(), "\"", "")
		jwtToken = strings.ReplaceAll(jwtToken, "\n", "")
		assert.Equal(t, harvesterTestSetup.JWTIssuer.Token, jwtToken)
	})
	t.Run("Onboard without join token fails", func(t *testing.T) {
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, "/onboard", "")
		echoCtx := harvesterTestSetup.EchoCtx

		SetupTrustDomain(t, harvesterTestSetup.Handler.Datastore)

		params := harvester.OnboardParams{
			JoinToken: "", // Empty join token
		}
		err := harvesterTestSetup.Handler.Onboard(echoCtx, params)
		require.Error(t, err)

		httpErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Equal(t, "join token is required", httpErr.Message)
	})
	t.Run("Onboard with join token that does not exist fails", func(t *testing.T) {
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, "/onboard", "")
		echoCtx := harvesterTestSetup.EchoCtx

		td := SetupTrustDomain(t, harvesterTestSetup.Handler.Datastore)
		SetupJoinToken(t, harvesterTestSetup.Handler.Datastore, td.ID.UUID)

		params := harvester.OnboardParams{
			JoinToken: "never-created-token",
		}
		err := harvesterTestSetup.Handler.Onboard(echoCtx, params)
		require.Error(t, err)

		httpErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Equal(t, "token not found", httpErr.Message)
	})
	t.Run("Onboard with join token that was used", func(t *testing.T) {
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, "/onboard", "")
		echoCtx := harvesterTestSetup.EchoCtx

		td := SetupTrustDomain(t, harvesterTestSetup.Handler.Datastore)
		token := SetupJoinToken(t, harvesterTestSetup.Handler.Datastore, td.ID.UUID)

		params := harvester.OnboardParams{
			JoinToken: token.Token,
		}
		err := harvesterTestSetup.Handler.Onboard(echoCtx, params)
		require.NoError(t, err)

		// repeat the request with the same token
		err = harvesterTestSetup.Handler.Onboard(echoCtx, params)
		require.Error(t, err)

		httpErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Equal(t, "token already used", httpErr.Message)
	})
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

		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodPut, "/trust-domain/:trustDomainName/bundles", string(body))
		echoCtx := harvesterTestSetup.EchoCtx

		// Creating Trust Domain
		td := SetupTrustDomain(t, harvesterTestSetup.Handler.Datastore)

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
