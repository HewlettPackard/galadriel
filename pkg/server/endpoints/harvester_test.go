package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/HewlettPackard/galadriel/test/fakes/fakejwtissuer"
	"github.com/HewlettPackard/galadriel/test/jwttest"
	gojwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	jwtPath           = "/jwt"
	onboardPath       = "/onboard"
	relationshipsPath = "/relationships"
)

var (
	tdA                   = &entity.TrustDomain{Name: spiffeid.RequireTrustDomainFromString("td-a.org"), ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	tdB                   = &entity.TrustDomain{Name: spiffeid.RequireTrustDomainFromString("td-b.org"), ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	tdC                   = &entity.TrustDomain{Name: spiffeid.RequireTrustDomainFromString("td-c.org"), ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	pendingRelAB          = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdB.ID.UUID, TrustDomainAConsent: entity.ConsentStatusPending, TrustDomainBConsent: entity.ConsentStatusPending, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	pendingRelAC          = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdC.ID.UUID, TrustDomainAConsent: entity.ConsentStatusPending, TrustDomainBConsent: entity.ConsentStatusPending, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	acceptedPendingRelAB  = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdB.ID.UUID, TrustDomainAConsent: entity.ConsentStatusAccepted, TrustDomainBConsent: entity.ConsentStatusPending, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	acceptedDeniedRelAC   = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdC.ID.UUID, TrustDomainAConsent: entity.ConsentStatusAccepted, TrustDomainBConsent: entity.ConsentStatusDenied, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	acceptedAcceptedRelBC = &entity.Relationship{TrustDomainAID: tdB.ID.UUID, TrustDomainBID: tdC.ID.UUID, TrustDomainAConsent: entity.ConsentStatusAccepted, TrustDomainBConsent: entity.ConsentStatusAccepted, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
)

type HarvesterTestSetup struct {
	EchoCtx   echo.Context
	Handler   *HarvesterAPIHandlers
	Datastore *datastore.FakeDatabase
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
		Datastore: fakeDB,
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

func SetupBundle(t *testing.T, ds datastore.Datastore, td uuid.UUID) *entity.Bundle {
	bundle := &entity.Bundle{
		TrustDomainID: td,
		Data:          []byte("test-bundle"),
		Signature:     []byte("test-signature"),
	}

	_, err := ds.CreateOrUpdateBundle(context.TODO(), bundle)
	require.NoError(t, err)

	return bundle
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
	t.Run("Successfully get accepted relationships", func(t *testing.T) {
		testGetRelationships(t, func(setup *HarvesterTestSetup, trustDomain *entity.TrustDomain) {
			setup.EchoCtx.Set(authTrustDomainKey, trustDomain)
		}, api.Accepted, tdA, 2)
	})

	t.Run("Successfully get denied relationships", func(t *testing.T) {
		testGetRelationships(t, func(setup *HarvesterTestSetup, trustDomain *entity.TrustDomain) {
			setup.EchoCtx.Set(authTrustDomainKey, trustDomain)
		}, api.Denied, tdC, 1)
	})

	t.Run("Successfully get pending relationships", func(t *testing.T) {
		testGetRelationships(t, func(setup *HarvesterTestSetup, trustDomain *entity.TrustDomain) {
			setup.EchoCtx.Set(authTrustDomainKey, trustDomain)
		}, api.Pending, tdB, 2)
	})

	t.Run("Successfully get all relationships", func(t *testing.T) {
		testGetRelationships(t, func(setup *HarvesterTestSetup, trustDomain *entity.TrustDomain) {
			setup.EchoCtx.Set(authTrustDomainKey, trustDomain)
		}, "", tdA, 4)
	})

	t.Run("Fails with invalid consent status", func(t *testing.T) {
		setup := NewHarvesterTestSetup(t, http.MethodGet, relationshipsPath, "")
		echoCtx := setup.EchoCtx
		setup.EchoCtx.Set(authTrustDomainKey, tdA)

		tdName := tdA.Name.String()
		status := api.ConsentStatus("invalid")
		params := harvester.GetRelationshipsParams{
			TrustDomainName: tdName,
			ConsentStatus:   &status,
		}

		err := setup.Handler.GetRelationships(echoCtx, params)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*echo.HTTPError).Code)
		assert.Equal(t, "invalid consent status: \"invalid\"", err.(*echo.HTTPError).Message)
	})

	t.Run("Fails if no authenticated trust domain", func(t *testing.T) {
		setup := NewHarvesterTestSetup(t, http.MethodGet, relationshipsPath, "")
		echoCtx := setup.EchoCtx

		tdName := tdA.Name.String()
		params := harvester.GetRelationshipsParams{
			TrustDomainName: tdName,
		}

		err := setup.Handler.GetRelationships(echoCtx, params)
		assert.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, err.(*echo.HTTPError).Code)
		assert.Equal(t, "no authenticated trust domain", err.(*echo.HTTPError).Message)
	})

	t.Run("Fails if authenticated trust domain does not match trust domain parameter", func(t *testing.T) {
		setup := NewHarvesterTestSetup(t, http.MethodGet, relationshipsPath, "")
		echoCtx := setup.EchoCtx
		setup.EchoCtx.Set(authTrustDomainKey, tdA)

		tdName := tdB.Name.String()
		params := harvester.GetRelationshipsParams{
			TrustDomainName: tdName,
		}

		err := setup.Handler.GetRelationships(echoCtx, params)
		assert.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, err.(*echo.HTTPError).Code)
		assert.Equal(t, "request trust domain \"td-b.org\" does not match authenticated trust domain \"td-a.org\"", err.(*echo.HTTPError).Message)
	})
}

func testGetRelationships(t *testing.T, setupFn func(*HarvesterTestSetup, *entity.TrustDomain), status api.ConsentStatus, trustDomain *entity.TrustDomain, expectedRelationshipCount int) {
	setup := NewHarvesterTestSetup(t, http.MethodGet, relationshipsPath, "")
	echoCtx := setup.EchoCtx

	setup.Datastore.WithTrustDomains(tdA, tdB, tdC)
	setup.Datastore.WithRelationships(pendingRelAB, pendingRelAC, acceptedPendingRelAB, acceptedDeniedRelAC, acceptedAcceptedRelBC)

	setupFn(setup, trustDomain)

	tdName := trustDomain.Name.String()
	params := harvester.GetRelationshipsParams{
		TrustDomainName: tdName,
		ConsentStatus:   &status,
	}

	err := setup.Handler.GetRelationships(echoCtx, params)
	assert.NoError(t, err)

	recorder := setup.Recorder
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.NotEmpty(t, recorder.Body)

	var relationships []*api.Relationship
	err = json.Unmarshal(recorder.Body.Bytes(), &relationships)
	assert.NoError(t, err)
	assert.Len(t, relationships, expectedRelationshipCount)

	if status == "" {
		return // no need to assert consent status
	}
	// assert that all relationships have the expected consent status for the specified trust domain
	for _, rel := range relationships {
		if rel.TrustDomainAId == trustDomain.ID.UUID {
			assert.Equal(t, status, rel.TrustDomainAConsent)
		}
		if rel.TrustDomainBId == trustDomain.ID.UUID {
			assert.Equal(t, status, rel.TrustDomainBConsent)
		}
	}
}

func TestTCPPatchRelationshipRelationshipID(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestTCPOnboard(t *testing.T) {
	t.Run("Successfully onboard a new agent", func(t *testing.T) {
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, onboardPath, "")
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
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, onboardPath, "")
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
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, onboardPath, "")
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
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, onboardPath, "")
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

func TestTCPGetNewJWTToken(t *testing.T) {
	t.Run("Successfully get a new JWT token", func(t *testing.T) {
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, jwtPath, "")
		echoCtx := harvesterTestSetup.EchoCtx

		SetupTrustDomain(t, harvesterTestSetup.Handler.Datastore)

		var claims gojwt.RegisteredClaims
		_, err := gojwt.ParseWithClaims(harvesterTestSetup.JWTIssuer.Token, &claims, func(*gojwt.Token) (interface{}, error) {
			return harvesterTestSetup.JWTIssuer.Signer.Public(), nil
		})
		assert.NoError(t, err)
		echoCtx.Set(authClaimsKey, &claims)

		err = harvesterTestSetup.Handler.GetNewJWTToken(echoCtx)
		assert.NoError(t, err)

		recorder := harvesterTestSetup.Recorder
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.NotEmpty(t, recorder.Body)

		jwtToken := strings.ReplaceAll(recorder.Body.String(), "\"", "")
		jwtToken = strings.ReplaceAll(jwtToken, "\n", "")
		assert.Equal(t, harvesterTestSetup.JWTIssuer.Token, jwtToken)
	})
	t.Run("Fails if no JWT token was sent", func(t *testing.T) {
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, jwtPath, "")
		echoCtx := harvesterTestSetup.EchoCtx

		err := harvesterTestSetup.Handler.GetNewJWTToken(echoCtx)
		require.Error(t, err)

		assert.Equal(t, http.StatusUnauthorized, err.(*echo.HTTPError).Code)
		assert.Equal(t, "failed to parse JWT access token claims", err.(*echo.HTTPError).Message)
	})
}

func TestTCPBundleSync(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestBundlePut(t *testing.T) {
	t.Run("Successfully post new bundle for a trust domain", func(t *testing.T) {
		setupFunc := func(setup *HarvesterTestSetup) *entity.TrustDomain {
			td := SetupTrustDomain(t, setup.Handler.Datastore)
			setup.EchoCtx.Set(authTrustDomainKey, td)
			return td
		}
		testBundlePut(t, setupFunc, http.StatusOK, "")
	})

	t.Run("Successfully post bundle update for a trust domain", func(t *testing.T) {
		setupFunc := func(setup *HarvesterTestSetup) *entity.TrustDomain {
			td := SetupTrustDomain(t, setup.Handler.Datastore)
			setup.EchoCtx.Set(authTrustDomainKey, td)
			SetupBundle(t, setup.Handler.Datastore, td.ID.UUID)
			return td
		}
		testBundlePut(t, setupFunc, http.StatusOK, "")
	})

	t.Run("Fail post bundle no authenticated trust domain", func(t *testing.T) {
		bundlePut := harvester.BundlePut{
			Signature:          "bundle signature",
			SigningCertificate: "certificate PEM",
			TrustBundle:        "a new bundle",
			TrustDomain:        testTrustDomain,
		}

		body, err := json.Marshal(bundlePut)
		require.NoError(t, err)

		setup := NewHarvesterTestSetup(t, http.MethodPut, "/trust-domain/:trustDomainName/bundles", string(body))
		setup.EchoCtx.Set(authTrustDomainKey, "")

		err = setup.Handler.BundlePut(setup.EchoCtx, testTrustDomain)
		require.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, err.(*echo.HTTPError).Code)
		assert.Equal(t, "no authenticated trust domain", err.(*echo.HTTPError).Message)
	})

	t.Run("Fail post bundle missing Trust bundle", func(t *testing.T) {
		testInvalidBundleRequest(t, "TrustBundle", "", http.StatusBadRequest, "invalid bundle request: trust bundle is required")
	})

	t.Run("Fail post bundle missing bundle signature", func(t *testing.T) {
		testInvalidBundleRequest(t, "Signature", "", http.StatusBadRequest, "invalid bundle request: bundle signature is required")
	})

	t.Run("Fail post bundle missing bundle trust domain", func(t *testing.T) {
		testInvalidBundleRequest(t, "TrustDomain", "", http.StatusBadRequest, "invalid bundle request: bundle trust domain is required")
	})

	t.Run("Fail post bundle bundle trust domain does not match authenticated trust domain", func(t *testing.T) {
		testInvalidBundleRequest(t, "TrustDomain", "other-trust-domain", http.StatusUnauthorized, "trust domain in request bundle \"other-trust-domain\" does not match authenticated trust domain: \"test.com\"")
	})
}

func testBundlePut(t *testing.T, setupFunc func(*HarvesterTestSetup) *entity.TrustDomain, expectedStatusCode int, expectedResponseBody string) {
	bundlePut := harvester.BundlePut{
		Signature:          "bundle signature",
		SigningCertificate: "certificate PEM",
		TrustBundle:        "a new bundle",
		TrustDomain:        testTrustDomain,
	}

	body, err := json.Marshal(bundlePut)
	require.NoError(t, err)

	setup := NewHarvesterTestSetup(t, http.MethodPut, "/trust-domain/:trustDomainName/bundles", string(body))
	echoCtx := setup.EchoCtx

	td := setupFunc(setup)

	err = setup.Handler.BundlePut(echoCtx, testTrustDomain)
	require.NoError(t, err)

	recorder := setup.Recorder
	assert.Equal(t, expectedStatusCode, recorder.Code)
	assert.Equal(t, expectedResponseBody, recorder.Body.String())

	storedBundle, err := setup.Handler.Datastore.FindBundleByTrustDomainID(context.Background(), td.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, bundlePut.Signature, string(storedBundle.Signature))
	assert.Equal(t, bundlePut.SigningCertificate, string(storedBundle.SigningCertificate))
	assert.Equal(t, bundlePut.TrustBundle, string(storedBundle.Data))
	assert.Equal(t, td.ID.UUID, storedBundle.TrustDomainID)
}

func testInvalidBundleRequest(t *testing.T, fieldName string, fieldValue interface{}, expectedStatusCode int, expectedErrorMessage string) {
	bundlePut := harvester.BundlePut{
		Signature:          "test-signature",
		SigningCertificate: "certificate PEM",
		TrustBundle:        "test trust bundle",
		TrustDomain:        testTrustDomain,
	}
	reflect.ValueOf(&bundlePut).Elem().FieldByName(fieldName).Set(reflect.ValueOf(fieldValue))

	body, err := json.Marshal(bundlePut)
	require.NoError(t, err)

	setup := NewHarvesterTestSetup(t, http.MethodPut, "/trust-domain/:trustDomainName/bundles", string(body))
	echoCtx := setup.EchoCtx

	td := SetupTrustDomain(t, setup.Handler.Datastore)
	echoCtx.Set(authTrustDomainKey, td)

	err = setup.Handler.BundlePut(echoCtx, testTrustDomain)
	require.Error(t, err)
	assert.Equal(t, expectedStatusCode, err.(*echo.HTTPError).Code)
	assert.Equal(t, expectedErrorMessage, err.(*echo.HTTPError).Message)
}
