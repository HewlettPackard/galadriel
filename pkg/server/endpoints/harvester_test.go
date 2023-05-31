package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/common/util/encoding"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/HewlettPackard/galadriel/pkg/server/db"
	"github.com/HewlettPackard/galadriel/test/fakes/fakedatastore"
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
	acceptedPendingRelAB  = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdB.ID.UUID, TrustDomainAConsent: entity.ConsentStatusApproved, TrustDomainBConsent: entity.ConsentStatusPending, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	deniedAcceptedRelAB   = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdB.ID.UUID, TrustDomainAConsent: entity.ConsentStatusDenied, TrustDomainBConsent: entity.ConsentStatusApproved, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	acceptedDeniedRelAC   = &entity.Relationship{TrustDomainAID: tdA.ID.UUID, TrustDomainBID: tdC.ID.UUID, TrustDomainAConsent: entity.ConsentStatusApproved, TrustDomainBConsent: entity.ConsentStatusDenied, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	acceptedAcceptedRelBC = &entity.Relationship{TrustDomainAID: tdB.ID.UUID, TrustDomainBID: tdC.ID.UUID, TrustDomainAConsent: entity.ConsentStatusApproved, TrustDomainBConsent: entity.ConsentStatusApproved, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}

	bundleA = &entity.Bundle{Data: []byte("bundle-A"), Digest: cryptoutil.CalculateDigest([]byte("bundle-A")), Signature: []byte("signature-A"), TrustDomainName: tdA.Name, TrustDomainID: tdA.ID.UUID, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	bundleB = &entity.Bundle{Data: []byte("bundle-B"), Digest: cryptoutil.CalculateDigest([]byte("bundle-B")), Signature: []byte("signature-B"), TrustDomainName: tdB.Name, TrustDomainID: tdB.ID.UUID, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
	bundleC = &entity.Bundle{Data: []byte("bundle-C"), Digest: cryptoutil.CalculateDigest([]byte("bundle-C")), Signature: []byte("signature-C"), TrustDomainName: tdC.Name, TrustDomainID: tdC.ID.UUID, ID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}
)

type HarvesterTestSetup struct {
	EchoCtx   echo.Context
	Handler   *HarvesterAPIHandlers
	Datastore *fakedatastore.FakeDatabase
	JWTIssuer *fakejwtissuer.JWTIssuer
	Recorder  *httptest.ResponseRecorder
}

func NewHarvesterTestSetup(t *testing.T, method, url string, body interface{}) *HarvesterTestSetup {
	var bodyReader io.Reader
	if body != nil {
		bodyStr, err := json.Marshal(body)
		assert.NoError(t, err)
		bodyReader = strings.NewReader(string(bodyStr))
	}

	e := echo.New()
	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	fakeDB := fakedatastore.NewFakeDB()
	logger := logrus.New()

	jwtAudience := []string{"test"}
	jwtIssuer := fakejwtissuer.New(t, "test", td1, jwtAudience)
	jwtValidator := jwttest.NewJWTValidator(jwtIssuer.Signer, jwtAudience)

	return &HarvesterTestSetup{
		EchoCtx:   e.NewContext(req, rec),
		Recorder:  rec,
		Handler:   NewHarvesterAPIHandlers(logger, fakeDB, jwtIssuer, jwtValidator),
		JWTIssuer: jwtIssuer,
		Datastore: fakeDB,
	}
}

func SetupTrustDomain(t *testing.T, ds db.Datastore) *entity.TrustDomain {
	td, err := spiffeid.TrustDomainFromString(td1)
	assert.NoError(t, err)

	tdEntity := &entity.TrustDomain{
		Name:        td,
		Description: "Fake domain",
	}
	trustDomain, err := ds.CreateOrUpdateTrustDomain(context.Background(), tdEntity)
	require.NoError(t, err)

	return trustDomain
}

func SetupBundle(t *testing.T, ds db.Datastore, td uuid.UUID) *entity.Bundle {
	bundle := &entity.Bundle{
		TrustDomainID: td,
		Data:          []byte("test-bundle"),
		Signature:     []byte("test-signature"),
	}

	_, err := ds.CreateOrUpdateBundle(context.Background(), bundle)
	require.NoError(t, err)

	return bundle
}

func SetupJoinToken(t *testing.T, ds db.Datastore, td uuid.UUID) *entity.JoinToken {
	jt := &entity.JoinToken{
		Token:         "test-join-token",
		TrustDomainID: td,
	}

	joinToken, err := ds.CreateJoinToken(context.Background(), jt)
	require.NoError(t, err)

	return joinToken
}

func TestTCPGetRelationships(t *testing.T) {
	t.Run("Successfully get approved relationships", func(t *testing.T) {
		testGetRelationships(t, func(setup *HarvesterTestSetup, trustDomain *entity.TrustDomain) {
			setup.EchoCtx.Set(authTrustDomainKey, trustDomain)
		}, api.Approved, tdA, 2)
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
		setup := NewHarvesterTestSetup(t, http.MethodGet, relationshipsPath, nil)
		echoCtx := setup.EchoCtx
		setup.EchoCtx.Set(authTrustDomainKey, tdA)

		tdName := tdA.Name.String()
		status := api.ConsentStatus("invalid")
		params := harvester.GetRelationshipsParams{
			ConsentStatus: &status,
		}

		err := setup.Handler.GetRelationships(echoCtx, tdName, params)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*echo.HTTPError).Code)
		assert.Contains(t, err.(*echo.HTTPError).Message, "invalid consent status: \"invalid\"")
	})

	t.Run("Fails if no authenticated trust domain", func(t *testing.T) {
		setup := NewHarvesterTestSetup(t, http.MethodGet, relationshipsPath, nil)
		echoCtx := setup.EchoCtx

		tdName := tdA.Name.String()
		params := harvester.GetRelationshipsParams{}

		err := setup.Handler.GetRelationships(echoCtx, tdName, params)
		assert.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, err.(*echo.HTTPError).Code)
		assert.Contains(t, err.(*echo.HTTPError).Message, "no authenticated trust domain")
	})

	t.Run("Fails if authenticated trust domain does not match trust domain parameter", func(t *testing.T) {
		setup := NewHarvesterTestSetup(t, http.MethodGet, relationshipsPath, nil)
		echoCtx := setup.EchoCtx
		setup.EchoCtx.Set(authTrustDomainKey, tdA)

		tdName := tdB.Name.String()
		params := harvester.GetRelationshipsParams{}

		err := setup.Handler.GetRelationships(echoCtx, tdName, params)
		assert.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, err.(*echo.HTTPError).Code)
		assert.Contains(t, err.(*echo.HTTPError).Message, "request trust domain \"td-b.org\" does not match authenticated trust domain \"td-a.org\"")
	})
}

func testGetRelationships(t *testing.T, setupFn func(*HarvesterTestSetup, *entity.TrustDomain), status api.ConsentStatus, trustDomain *entity.TrustDomain, expectedRelationshipCount int) {
	setup := NewHarvesterTestSetup(t, http.MethodGet, relationshipsPath, nil)
	echoCtx := setup.EchoCtx

	setup.Datastore.WithTrustDomains(tdA, tdB, tdC)
	setup.Datastore.WithRelationships(pendingRelAB, pendingRelAC, acceptedPendingRelAB, acceptedDeniedRelAC, acceptedAcceptedRelBC)

	setupFn(setup, trustDomain)

	tdName := trustDomain.Name.String()
	params := harvester.GetRelationshipsParams{
		ConsentStatus: &status,
	}

	err := setup.Handler.GetRelationships(echoCtx, tdName, params)
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
	t.Run("Successfully patch pending relationship to approved", func(t *testing.T) {
		testPatchRelationship(t, func(setup *HarvesterTestSetup, trustDomain *entity.TrustDomain) {
			setup.EchoCtx.Set(authTrustDomainKey, tdA)
		}, tdA, pendingRelAC, api.Approved)
	})
	t.Run("Successfully patch pending relationship to denied", func(t *testing.T) {
		testPatchRelationship(t, func(setup *HarvesterTestSetup, trustDomain *entity.TrustDomain) {
			setup.EchoCtx.Set(authTrustDomainKey, tdA)
		}, tdA, pendingRelAC, api.Denied)
	})
	t.Run("Successfully patch pending relationship to approved with other trust domain", func(t *testing.T) {
		testPatchRelationship(t, func(setup *HarvesterTestSetup, trustDomain *entity.TrustDomain) {
			setup.EchoCtx.Set(authTrustDomainKey, tdC)
		}, tdC, pendingRelAC, api.Approved)
	})
	t.Run("Successfully patch approved relationship to denied", func(t *testing.T) {
		testPatchRelationship(t, func(setup *HarvesterTestSetup, trustDomain *entity.TrustDomain) {
			setup.EchoCtx.Set(authTrustDomainKey, tdB)
		}, tdB, acceptedAcceptedRelBC, api.Denied)
	})
}

func testPatchRelationship(t *testing.T, f func(setup *HarvesterTestSetup, trustDomain *entity.TrustDomain), trustDomain *entity.TrustDomain, relationship *entity.Relationship, status api.ConsentStatus) {
	requestBody := &harvester.PatchRelationshipRequest{
		ConsentStatus: status,
	}

	setup := NewHarvesterTestSetup(t, http.MethodPatch, relationshipsPath+"/"+relationship.ID.UUID.String(), &requestBody)
	echoCtx := setup.EchoCtx

	setup.Datastore.WithTrustDomains(tdA, tdB, tdC)
	setup.Datastore.WithRelationships(relationship)

	f(setup, trustDomain)

	err := setup.Handler.PatchRelationship(echoCtx, trustDomain.Name.String(), relationship.ID.UUID)
	assert.NoError(t, err)

	recorder := setup.Recorder
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, status, status)

	// lookup relationship to assert that it was updated
	rel, err := setup.Datastore.FindRelationshipByID(context.Background(), relationship.ID.UUID)
	assert.NoError(t, err)

	if rel.TrustDomainAID == trustDomain.ID.UUID {
		assert.Equal(t, entity.ConsentStatus(status), rel.TrustDomainAConsent)
		// the other trust domain's consent status should not have changed
		assert.Equal(t, relationship.TrustDomainBConsent, rel.TrustDomainBConsent)
	} else {
		assert.Equal(t, entity.ConsentStatus(status), rel.TrustDomainBConsent)
		// the other trust domain's consent status should not have changed
		assert.Equal(t, relationship.TrustDomainAConsent, rel.TrustDomainAConsent)
	}

	var resp api.Relationship
	err = json.Unmarshal(recorder.Body.Bytes(), &resp)
	expected := api.MapRelationships(relationship)[0]
	require.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.Equal(t, expected.Id, resp.Id)
	assert.Equal(t, expected.TrustDomainAId, resp.TrustDomainAId)
	assert.Equal(t, expected.TrustDomainBId, resp.TrustDomainBId)
	assert.Equal(t, expected.TrustDomainAName, resp.TrustDomainAName)
	assert.Equal(t, expected.TrustDomainBName, resp.TrustDomainBName)
	assert.Equal(t, expected.TrustDomainAConsent, resp.TrustDomainAConsent)
}

func TestTCPOnboard(t *testing.T) {
	t.Run("Successfully onboard a new agent", func(t *testing.T) {
		// Arrange
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, onboardPath, nil)
		echoCtx := harvesterTestSetup.EchoCtx

		td := SetupTrustDomain(t, harvesterTestSetup.Handler.Datastore)
		token := SetupJoinToken(t, harvesterTestSetup.Handler.Datastore, td.ID.UUID)

		params := harvester.OnboardParams{
			JoinToken: token.Token,
		}

		// Act
		err := harvesterTestSetup.Handler.Onboard(echoCtx, td.Name.String(), params)
		assert.NoError(t, err)

		// Assert
		recorder := harvesterTestSetup.Recorder
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.NotEmpty(t, recorder.Body)

		var result harvester.OnboardHarvesterResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, td.ID.UUID.String(), result.TrustDomainID.String())
		assert.Equal(t, td.Name.String(), result.TrustDomainName)

		assert.NotEmpty(t, result.Token)
		jwtToken := strings.ReplaceAll(result.Token, "\"", "")
		jwtToken = strings.ReplaceAll(jwtToken, "\n", "")
		assert.Equal(t, harvesterTestSetup.JWTIssuer.Token, jwtToken)
	})
	t.Run("onboard without join token fails", func(t *testing.T) {
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, onboardPath, nil)
		echoCtx := harvesterTestSetup.EchoCtx

		td := SetupTrustDomain(t, harvesterTestSetup.Handler.Datastore)

		params := harvester.OnboardParams{
			JoinToken: "", // Empty join token
		}
		err := harvesterTestSetup.Handler.Onboard(echoCtx, td.Name.String(), params)
		require.Error(t, err)

		httpErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Contains(t, httpErr.Message, "join token is required")
	})
	t.Run("onboard with join token that does not exist fails", func(t *testing.T) {
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, onboardPath, nil)
		echoCtx := harvesterTestSetup.EchoCtx

		td := SetupTrustDomain(t, harvesterTestSetup.Handler.Datastore)
		SetupJoinToken(t, harvesterTestSetup.Handler.Datastore, td.ID.UUID)

		params := harvester.OnboardParams{
			JoinToken: "never-created-token",
		}
		err := harvesterTestSetup.Handler.Onboard(echoCtx, td.Name.String(), params)
		require.Error(t, err)

		httpErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Contains(t, httpErr.Message, "token not found")
	})
	t.Run("onboard with join token that was used", func(t *testing.T) {
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, onboardPath, nil)
		echoCtx := harvesterTestSetup.EchoCtx

		td := SetupTrustDomain(t, harvesterTestSetup.Handler.Datastore)
		token := SetupJoinToken(t, harvesterTestSetup.Handler.Datastore, td.ID.UUID)

		params := harvester.OnboardParams{
			JoinToken: token.Token,
		}
		err := harvesterTestSetup.Handler.Onboard(echoCtx, td.Name.String(), params)
		require.NoError(t, err)

		// repeat the request with the same token
		err = harvesterTestSetup.Handler.Onboard(echoCtx, td.Name.String(), params)
		require.Error(t, err)

		httpErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Contains(t, httpErr.Message, "token already used")
	})
}

func TestTCPGetNewJWTToken(t *testing.T) {
	t.Run("Successfully get a new JWT token", func(t *testing.T) {
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, jwtPath, nil)
		echoCtx := harvesterTestSetup.EchoCtx

		td := SetupTrustDomain(t, harvesterTestSetup.Handler.Datastore)

		var claims gojwt.RegisteredClaims
		_, err := gojwt.ParseWithClaims(harvesterTestSetup.JWTIssuer.Token, &claims, func(*gojwt.Token) (interface{}, error) {
			return harvesterTestSetup.JWTIssuer.Signer.Public(), nil
		})
		assert.NoError(t, err)
		echoCtx.Set(authClaimsKey, &claims)

		err = harvesterTestSetup.Handler.GetNewJWTToken(echoCtx, td.Name.String())
		assert.NoError(t, err)

		recorder := harvesterTestSetup.Recorder
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.NotEmpty(t, recorder.Body)

		jwtToken := strings.ReplaceAll(recorder.Body.String(), "\"", "")
		jwtToken = strings.ReplaceAll(jwtToken, "\n", "")
		expected := fmt.Sprintf("{token:%s}", harvesterTestSetup.JWTIssuer.Token)
		assert.Equal(t, expected, jwtToken)
	})
	t.Run("Fails if no JWT token was sent", func(t *testing.T) {
		harvesterTestSetup := NewHarvesterTestSetup(t, http.MethodGet, jwtPath, nil)
		echoCtx := harvesterTestSetup.EchoCtx

		err := harvesterTestSetup.Handler.GetNewJWTToken(echoCtx, "td1")
		require.Error(t, err)

		assert.Equal(t, http.StatusUnauthorized, err.(*echo.HTTPError).Code)
		assert.Contains(t, err.(*echo.HTTPError).Message, "failed to parse JWT access token claims")
	})
}

func TestTCPBundleSync(t *testing.T) {
	testCases := []struct {
		name          string
		trustDomain   string
		relationships []*entity.Relationship
		bundleState   harvester.PostBundleSyncRequest
		expected      harvester.PostBundleSyncResponse
	}{
		{
			name:          "Successfully sync no new bundles",
			trustDomain:   tdA.Name.String(),
			relationships: []*entity.Relationship{acceptedPendingRelAB, acceptedDeniedRelAC, acceptedAcceptedRelBC},
			bundleState: harvester.PostBundleSyncRequest{
				State: map[string]api.BundleDigest{
					tdB.Name.String(): encoding.EncodeToBase64(bundleB.Digest),
					tdC.Name.String(): encoding.EncodeToBase64(bundleC.Digest),
				},
			},
			expected: harvester.PostBundleSyncResponse{
				State: harvester.BundlesDigests{
					tdB.Name.String(): encoding.EncodeToBase64(bundleB.Digest),
					tdC.Name.String(): encoding.EncodeToBase64(bundleC.Digest),
				},
				Updates: harvester.BundlesUpdates{},
			},
		},
		{
			name:          "Successfully sync one new bundle for one approved relationship",
			trustDomain:   tdA.Name.String(),
			relationships: []*entity.Relationship{acceptedPendingRelAB, acceptedDeniedRelAC, acceptedAcceptedRelBC},
			bundleState: harvester.PostBundleSyncRequest{
				State: map[string]api.BundleDigest{
					tdC.Name.String(): encoding.EncodeToBase64(bundleC.Digest),
				},
			},
			expected: harvester.PostBundleSyncResponse{
				State: harvester.BundlesDigests{
					tdB.Name.String(): encoding.EncodeToBase64(bundleB.Digest),
					tdC.Name.String(): encoding.EncodeToBase64(bundleC.Digest),
				},
				Updates: harvester.BundlesUpdates{
					tdB.Name.String(): harvester.BundlesUpdatesItem{
						TrustBundle: string(bundleB.Data),
						Digest:      encoding.EncodeToBase64(bundleB.Digest),
						Signature:   encoding.EncodeToBase64(bundleB.Signature),
					},
				},
			},
		},
		{
			name:          "Successfully sync two new bundles for two approved relationships",
			trustDomain:   tdA.Name.String(),
			relationships: []*entity.Relationship{acceptedPendingRelAB, acceptedDeniedRelAC, acceptedAcceptedRelBC},
			bundleState: harvester.PostBundleSyncRequest{
				State: map[string]api.BundleDigest{},
			},
			expected: harvester.PostBundleSyncResponse{
				State: harvester.BundlesDigests{
					tdB.Name.String(): encoding.EncodeToBase64(bundleB.Digest),
					tdC.Name.String(): encoding.EncodeToBase64(bundleC.Digest),
				},
				Updates: harvester.BundlesUpdates{
					tdB.Name.String(): harvester.BundlesUpdatesItem{
						TrustBundle: string(bundleB.Data),
						Digest:      encoding.EncodeToBase64(bundleB.Digest),
						Signature:   encoding.EncodeToBase64(bundleB.Signature),
					},
					tdC.Name.String(): harvester.BundlesUpdatesItem{
						TrustBundle: string(bundleC.Data),
						Digest:      encoding.EncodeToBase64(bundleC.Digest),
						Signature:   encoding.EncodeToBase64(bundleC.Signature),
					},
				},
			},
		},
		{
			name:          "Successfully sync one new bundle for one approved relationship, not including the pending relationship",
			trustDomain:   tdA.Name.String(),
			relationships: []*entity.Relationship{acceptedPendingRelAB, pendingRelAC, acceptedAcceptedRelBC},
			bundleState: harvester.PostBundleSyncRequest{
				State: map[string]api.BundleDigest{},
			},
			expected: harvester.PostBundleSyncResponse{
				State: harvester.BundlesDigests{
					tdB.Name.String(): encoding.EncodeToBase64(bundleB.Digest),
				},
				Updates: harvester.BundlesUpdates{
					tdB.Name.String(): harvester.BundlesUpdatesItem{
						TrustBundle: string(bundleB.Data),
						Digest:      encoding.EncodeToBase64(bundleB.Digest),
						Signature:   encoding.EncodeToBase64(bundleB.Signature),
					},
				},
			},
		},
		{
			name:          "Successfully sync one new bundle for one approved relationship, not including the denied relationship",
			trustDomain:   tdA.Name.String(),
			relationships: []*entity.Relationship{acceptedDeniedRelAC, deniedAcceptedRelAB, acceptedAcceptedRelBC},
			bundleState: harvester.PostBundleSyncRequest{
				State: map[string]api.BundleDigest{},
			},
			expected: harvester.PostBundleSyncResponse{
				State: harvester.BundlesDigests{
					tdC.Name.String(): encoding.EncodeToBase64(bundleC.Digest),
				},
				Updates: harvester.BundlesUpdates{
					tdC.Name.String(): harvester.BundlesUpdatesItem{
						TrustBundle: string(bundleC.Data),
						Digest:      encoding.EncodeToBase64(bundleC.Digest),
						Signature:   encoding.EncodeToBase64(bundleC.Signature),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup := NewHarvesterTestSetup(t, http.MethodPost, "/trust-domain/:trustDomainName/bundles/sync", &tc.bundleState)
			echoCtx := setup.EchoCtx
			setup.EchoCtx.Set(authTrustDomainKey, tdA)

			setup.Datastore.WithTrustDomains(tdA, tdB, tdC)
			setup.Datastore.WithRelationships(tc.relationships...)
			setup.Datastore.WithBundles(bundleA, bundleB, bundleC)

			// test bundle sync
			err := setup.Handler.BundleSync(echoCtx, tdA.Name.String())
			assert.NoError(t, err)

			recorder := setup.Recorder
			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.NoError(t, err)

			var bundles harvester.PostBundleSyncResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &bundles)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, bundles)
		})
	}
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
		sig := "test-signature"
		cert := "test-certificate"
		bundlePut := harvester.PutBundleRequest{
			Signature:          &sig,
			SigningCertificate: &cert,
			TrustBundle:        "a new bundle",
			TrustDomain:        td1,
		}

		setup := NewHarvesterTestSetup(t, http.MethodPut, "/trust-domain/:trustDomainName/bundles", &bundlePut)
		setup.EchoCtx.Set(authTrustDomainKey, "")

		err := setup.Handler.BundlePut(setup.EchoCtx, td1)
		require.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, err.(*echo.HTTPError).Code)
		assert.Contains(t, err.(*echo.HTTPError).Message, "no authenticated trust domain")
	})

	t.Run("Fail post bundle missing Trust bundle", func(t *testing.T) {
		testInvalidBundleRequest(t, "TrustBundle", "", http.StatusBadRequest, "invalid bundle request: trust bundle is required")
	})

	t.Run("Fail post bundle missing bundle trust domain", func(t *testing.T) {
		testInvalidBundleRequest(t, "TrustDomain", "", http.StatusBadRequest, "invalid bundle request: bundle trust domain is required")
	})

	t.Run("Fail post bundle trust domain does not match authenticated trust domain", func(t *testing.T) {
		testInvalidBundleRequest(t, "TrustDomain", "other-trust-domain", http.StatusUnauthorized, "trust domain in request bundle \"other-trust-domain\" does not match authenticated trust domain: \"test1.com\"")
	})
}

func testBundlePut(t *testing.T, setupFunc func(*HarvesterTestSetup) *entity.TrustDomain, expectedStatusCode int, expectedResponseBody string) {
	bundle := "a new bundle"
	digest := encoding.EncodeToBase64(cryptoutil.CalculateDigest([]byte(bundle)))
	sig := encoding.EncodeToBase64([]byte("test-signature"))
	cert := encoding.EncodeToBase64([]byte("test-signing-certificate"))
	bundlePut := harvester.PutBundleRequest{
		Signature:          &sig,
		SigningCertificate: &cert,
		TrustBundle:        bundle,
		Digest:             digest,
		TrustDomain:        td1,
	}

	setup := NewHarvesterTestSetup(t, http.MethodPut, "/trust-domain/:trustDomainName/bundles", &bundlePut)
	echoCtx := setup.EchoCtx

	td := setupFunc(setup)

	err := setup.Handler.BundlePut(echoCtx, td1)
	require.NoError(t, err)

	recorder := setup.Recorder
	assert.Equal(t, expectedStatusCode, recorder.Code)
	assert.Equal(t, expectedResponseBody, recorder.Body.String())

	storedBundle, err := setup.Handler.Datastore.FindBundleByTrustDomainID(context.Background(), td.ID.UUID)
	require.NoError(t, err)
	assert.Equal(t, bundlePut.TrustBundle, string(storedBundle.Data))
	assert.Equal(t, digest, encoding.EncodeToBase64(storedBundle.Digest))
	assert.Equal(t, sig, encoding.EncodeToBase64(storedBundle.Signature))
	assert.Equal(t, cert, encoding.EncodeToBase64(storedBundle.SigningCertificate))
	assert.Equal(t, td.ID.UUID, storedBundle.TrustDomainID)
}

func testInvalidBundleRequest(t *testing.T, fieldName string, fieldValue interface{}, expectedStatusCode int, expectedErrorMessage string) {
	sig := "test-signature"
	cert := "test-certificate"
	bundle := "test trust bundle"
	digest := encoding.EncodeToBase64(cryptoutil.CalculateDigest([]byte(bundle)))
	bundlePut := harvester.PutBundleRequest{
		Signature:          &sig,
		SigningCertificate: &cert,
		TrustBundle:        bundle,
		Digest:             digest,
		TrustDomain:        td1,
	}
	reflect.ValueOf(&bundlePut).Elem().FieldByName(fieldName).Set(reflect.ValueOf(fieldValue))

	setup := NewHarvesterTestSetup(t, http.MethodPut, "/trust-domain/:trustDomainName/bundles", &bundlePut)
	echoCtx := setup.EchoCtx

	td := SetupTrustDomain(t, setup.Handler.Datastore)
	echoCtx.Set(authTrustDomainKey, td)

	err := setup.Handler.BundlePut(echoCtx, td1)
	require.Error(t, err)
	assert.Equal(t, expectedStatusCode, err.(*echo.HTTPError).Code)
	assert.Contains(t, err.(*echo.HTTPError).Message, expectedErrorMessage)
}
