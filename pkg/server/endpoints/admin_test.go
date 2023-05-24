package endpoints

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/api/admin"
	"github.com/HewlettPackard/galadriel/test/fakes/fakedatastore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Trust Domains
	td1 = "test1.com"
	td2 = "test2.com"
	td3 = "test3.com"
)

var (
	// Relationships ID's
	r1ID = NewNullableID()
	r2ID = NewNullableID()
	r3ID = NewNullableID()
	r4ID = NewNullableID()
	r5ID = NewNullableID()

	// Trust Domains ID's
	tdUUID1 = NewNullableID()
	tdUUID2 = NewNullableID()
	tdUUID3 = NewNullableID()

	spiffeTD1    = spiffeid.RequireTrustDomainFromString(td1)
	spiffeTD2    = spiffeid.RequireTrustDomainFromString(td2)
	spiffeTD3    = spiffeid.RequireTrustDomainFromString(td3)
	entTD1       = &entity.TrustDomain{ID: tdUUID1, Name: spiffeTD1}
	entTD2       = &entity.TrustDomain{ID: tdUUID2, Name: spiffeTD2}
	entTD3       = &entity.TrustDomain{ID: tdUUID3, Name: spiffeTD3}
	trustDomains = []*entity.TrustDomain{entTD1, entTD2, entTD3}

	rel1          = &entity.Relationship{ID: r1ID, TrustDomainAID: entTD1.ID.UUID, TrustDomainBID: tdUUID2.UUID, TrustDomainAName: spiffeTD1, TrustDomainBName: spiffeTD2, TrustDomainAConsent: entity.ConsentStatus(api.Approved), TrustDomainBConsent: entity.ConsentStatus(api.Pending)}
	rel2          = &entity.Relationship{ID: r2ID, TrustDomainAID: entTD1.ID.UUID, TrustDomainBID: tdUUID3.UUID, TrustDomainAName: spiffeTD1, TrustDomainBName: spiffeTD3, TrustDomainAConsent: entity.ConsentStatus(api.Denied), TrustDomainBConsent: entity.ConsentStatus(api.Approved)}
	rel3          = &entity.Relationship{ID: r3ID, TrustDomainAID: entTD2.ID.UUID, TrustDomainBID: tdUUID3.UUID, TrustDomainAName: spiffeTD2, TrustDomainBName: spiffeTD3, TrustDomainAConsent: entity.ConsentStatus(api.Approved), TrustDomainBConsent: entity.ConsentStatus(api.Denied)}
	rel4          = &entity.Relationship{ID: r4ID, TrustDomainAID: entTD3.ID.UUID, TrustDomainBID: tdUUID1.UUID, TrustDomainAName: spiffeTD3, TrustDomainBName: spiffeTD1, TrustDomainAConsent: entity.ConsentStatus(api.Denied), TrustDomainBConsent: entity.ConsentStatus(api.Denied)}
	rel5          = &entity.Relationship{ID: r5ID, TrustDomainAID: entTD3.ID.UUID, TrustDomainBID: tdUUID2.UUID, TrustDomainAName: spiffeTD3, TrustDomainBName: spiffeTD2, TrustDomainAConsent: entity.ConsentStatus(api.Pending), TrustDomainBConsent: entity.ConsentStatus(api.Pending)}
	relationships = []*entity.Relationship{rel1, rel2, rel3, rel4, rel5}
)

type ManagementTestSetup struct {
	EchoCtx      echo.Context
	Handler      *AdminAPIHandlers
	Recorder     *httptest.ResponseRecorder
	FakeDatabase *fakedatastore.FakeDatabase

	// Helpers
	bodyReader io.Reader

	url    string
	method string
}

func NewManagementTestSetup(t *testing.T, method, url string, body interface{}) *ManagementTestSetup {
	var bodyReader io.Reader = nil
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

	return &ManagementTestSetup{
		EchoCtx:      e.NewContext(req, rec),
		Recorder:     rec,
		Handler:      NewAdminAPIHandlers(logger, fakeDB),
		FakeDatabase: fakeDB,
		// Helpers
		url:        url,
		method:     method,
		bodyReader: bodyReader,
	}
}

func (setup *ManagementTestSetup) Refresh() {
	e := echo.New()
	req := httptest.NewRequest(setup.method, setup.url, setup.bodyReader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	// Refreshing Request context and Recorder
	setup.EchoCtx = e.NewContext(req, rec)
	setup.Recorder = rec
}

func TestGetRelationships(t *testing.T) {
	tdName := td1
	statusAccepted := api.Approved
	statusPending := api.Pending
	statusDenied := api.Denied

	t.Run("Successfully filter by trust domain", func(t *testing.T) {
		runGetRelationshipTest(t, admin.GetRelationshipsParams{TrustDomainName: &tdName}, 3, rel1, rel2, rel4)
	})

	t.Run("Successfully filter by status approved", func(t *testing.T) {
		runGetRelationshipTest(t, admin.GetRelationshipsParams{Status: &statusAccepted}, 3, rel1, rel2, rel3)
	})

	t.Run("Successfully filter by status pending", func(t *testing.T) {
		runGetRelationshipTest(t, admin.GetRelationshipsParams{Status: &statusPending}, 2, rel1, rel5)
	})

	t.Run("Successfully filter by status denied", func(t *testing.T) {
		runGetRelationshipTest(t, admin.GetRelationshipsParams{Status: &statusDenied}, 3, rel2, rel3, rel4)
	})

	t.Run("Successfully filter by status approved and trust domain", func(t *testing.T) {
		runGetRelationshipTest(t, admin.GetRelationshipsParams{TrustDomainName: &tdName, Status: &statusAccepted}, 1, rel1)
	})

	t.Run("Should raise a bad request when receiving undefined status filter", func(t *testing.T) {
		// Setup
		setup := NewManagementTestSetup(t, http.MethodGet, relationshipsPath, nil)

		// Approved filter
		var randomFilter api.ConsentStatus = "a random filter"
		params := admin.GetRelationshipsParams{
			Status: &randomFilter,
		}

		err := setup.Handler.GetRelationships(setup.EchoCtx, params)
		assert.Error(t, err)

		httpErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Empty(t, setup.Recorder.Body)

		expectedMsg := fmt.Sprintf(
			"status filter %q is not supported, approved values [%v, %v, %v]",
			randomFilter, api.Approved, api.Denied, api.Pending,
		)

		assert.ErrorContains(t, err, expectedMsg)
	})
}

func runGetRelationshipTest(t *testing.T, params admin.GetRelationshipsParams, expectedLength int, expectedRelationships ...*entity.Relationship) {
	setup := setupGetRelationshipTest(t)

	err := setup.Handler.GetRelationships(setup.EchoCtx, params)
	assert.NoError(t, err)

	recorder := setup.Recorder
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.NotEmpty(t, recorder.Body)

	var relationships []*api.Relationship
	err = json.Unmarshal(recorder.Body.Bytes(), &relationships)
	assert.NoError(t, err)

	assert.Equal(t, expectedLength, len(relationships))
	assertContainRelationships(t, relationships, api.MapRelationships(expectedRelationships...))
}

func setupGetRelationshipTest(t *testing.T) *ManagementTestSetup {
	managementTestSetup := NewManagementTestSetup(t, http.MethodGet, "/relationships", nil)

	managementTestSetup.FakeDatabase.WithTrustDomains(trustDomains...)
	managementTestSetup.FakeDatabase.WithRelationships(relationships...)

	return managementTestSetup
}

func TestUDSPutRelationships(t *testing.T) {
	relationshipsPath := "/relationships"

	t.Run("Successfully create a new relationship request", func(t *testing.T) {

		fakeTrustDomains := []*entity.TrustDomain{
			{ID: tdUUID1, Name: NewTrustDomain(t, td1)},
			{ID: tdUUID2, Name: NewTrustDomain(t, td2)},
		}

		reqBody := &admin.PutRelationshipJSONRequestBody{
			TrustDomainAName: td1,
			TrustDomainBName: td2,
		}

		// Setup
		setup := NewManagementTestSetup(t, http.MethodPut, relationshipsPath, reqBody)
		setup.FakeDatabase.WithTrustDomains(fakeTrustDomains...)

		err := setup.Handler.PutRelationship(setup.EchoCtx)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusCreated, setup.Recorder.Code)

		apiRelation := api.Relationship{}
		err = json.Unmarshal(setup.Recorder.Body.Bytes(), &apiRelation)
		assert.NoError(t, err)

		assert.NotNil(t, apiRelation)
		assert.Equal(t, tdUUID1.UUID, apiRelation.TrustDomainAId)
		assert.Equal(t, tdUUID2.UUID, apiRelation.TrustDomainBId)
	})

	t.Run("Should not allow relationships request between inexistent trust domains", func(t *testing.T) {

		fakeTrustDomains := []*entity.TrustDomain{
			{ID: tdUUID1, Name: NewTrustDomain(t, td1)},
		}

		reqBody := &admin.PutRelationshipJSONRequestBody{
			TrustDomainAName: td1,
			TrustDomainBName: td2,
		}

		// Setup
		setup := NewManagementTestSetup(t, http.MethodPut, relationshipsPath, reqBody)
		setup.FakeDatabase.WithTrustDomains(fakeTrustDomains...)

		err := setup.Handler.PutRelationship(setup.EchoCtx)
		assert.Error(t, err)
		assert.Empty(t, setup.Recorder.Body.Bytes())

		echoHttpErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusNotFound, echoHttpErr.Code)

		expectedErrorMsg := fmt.Sprintf("trust domain does not exist: %q", td2)
		assert.Equal(t, expectedErrorMsg, echoHttpErr.Message)
	})

	// Should we test sending wrong body formats ?
}

func TestUDSGetRelationshipsByID(t *testing.T) {
	relationshipsPath := "/relationships/%v"

	t.Run("Successfully get relationship information", func(t *testing.T) {

		fakeTrustDomains := []*entity.TrustDomain{
			{ID: tdUUID1, Name: NewTrustDomain(t, td1)},
			{ID: tdUUID2, Name: NewTrustDomain(t, td2)},
		}

		r1ID := NewNullableID()
		fakeRelationship := &entity.Relationship{
			ID:             r1ID,
			TrustDomainAID: tdUUID1.UUID,
			TrustDomainBID: tdUUID2.UUID,
		}

		completePath := fmt.Sprintf(relationshipsPath, r1ID.UUID)

		// Setup
		setup := NewManagementTestSetup(t, http.MethodGet, completePath, nil)
		setup.FakeDatabase.WithTrustDomains(fakeTrustDomains...)
		setup.FakeDatabase.WithRelationships(fakeRelationship)

		err := setup.Handler.GetRelationshipByID(setup.EchoCtx, r1ID.UUID)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, setup.Recorder.Code)

		apiRelation := api.Relationship{}
		err = json.Unmarshal(setup.Recorder.Body.Bytes(), &apiRelation)
		assert.NoError(t, err)

		assert.NotNil(t, apiRelation)
		assert.Equal(t, tdUUID1.UUID, apiRelation.TrustDomainAId)
		assert.Equal(t, tdUUID2.UUID, apiRelation.TrustDomainBId)
	})

	t.Run("Should raise a not found request when try to get information about a relationship that doesn't exists", func(t *testing.T) {
		completePath := fmt.Sprintf(relationshipsPath, r1ID.UUID)

		// Setup
		setup := NewManagementTestSetup(t, http.MethodGet, completePath, nil)

		err := setup.Handler.GetRelationshipByID(setup.EchoCtx, r1ID.UUID)
		assert.Error(t, err)
		assert.Empty(t, setup.Recorder.Body.Bytes())

		echoHTTPerr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusNotFound, echoHTTPerr.Code)
		assert.Equal(t, "relationship not found", echoHTTPerr.Message)
	})
}

func TestUDSPutTrustDomain(t *testing.T) {
	trustDomainPath := "/trust-domain"
	t.Run("Successfully create a new trust domain", func(t *testing.T) {
		description := "A test trust domain"
		reqBody := &admin.PutTrustDomainRequest{
			Name:        td1,
			Description: &description,
		}

		// Setup
		setup := NewManagementTestSetup(t, http.MethodPut, trustDomainPath, reqBody)

		err := setup.Handler.PutTrustDomain(setup.EchoCtx)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusCreated, setup.Recorder.Code)

		apiTrustDomain := api.TrustDomain{}
		err = json.Unmarshal(setup.Recorder.Body.Bytes(), &apiTrustDomain)
		assert.NoError(t, err)

		assert.NotNil(t, apiTrustDomain)
		assert.Equal(t, td1, apiTrustDomain.Name)
		assert.Equal(t, description, *apiTrustDomain.Description)

	})

	t.Run("Should not allow creating trust domain with the same name of one already created", func(t *testing.T) {
		reqBody := &admin.PutTrustDomainRequest{
			Name: td1,
		}

		fakeTrustDomains := entity.TrustDomain{ID: NewNullableID(), Name: NewTrustDomain(t, td1)}

		// Setup
		setup := NewManagementTestSetup(t, http.MethodPut, trustDomainPath, reqBody)
		setup.FakeDatabase.WithTrustDomains(&fakeTrustDomains)

		err := setup.Handler.PutTrustDomain(setup.EchoCtx)
		assert.Error(t, err)

		echoHttpErr := err.(*echo.HTTPError)

		assert.Equal(t, http.StatusBadRequest, echoHttpErr.Code)
		expectedErrorMsg := fmt.Sprintf("trust domain already exists: %q", td1)
		assert.Equal(t, expectedErrorMsg, echoHttpErr.Message)
	})
}

func TestUDSGetTrustDomainByName(t *testing.T) {
	trustDomainPath := "/trust-domain/%v"

	t.Run("Successfully retrieve trust domain information", func(t *testing.T) {
		fakeTrustDomains := entity.TrustDomain{ID: tdUUID1, Name: NewTrustDomain(t, td1)}

		completePath := fmt.Sprintf(trustDomainPath, tdUUID1.UUID)

		// Setup
		setup := NewManagementTestSetup(t, http.MethodPut, completePath, nil)
		setup.FakeDatabase.WithTrustDomains(&fakeTrustDomains)

		err := setup.Handler.GetTrustDomainByName(setup.EchoCtx, td1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, setup.Recorder.Code)

		apiTrustDomain := api.TrustDomain{}
		err = json.Unmarshal(setup.Recorder.Body.Bytes(), &apiTrustDomain)
		assert.NoError(t, err)

		assert.Equal(t, td1, apiTrustDomain.Name)
		assert.Equal(t, tdUUID1.UUID, apiTrustDomain.Id)
	})

	t.Run("Raise a not found when trying to retrieve a trust domain that does not exist", func(t *testing.T) {
		completePath := fmt.Sprintf(trustDomainPath, tdUUID1.UUID)

		// Setup
		setup := NewManagementTestSetup(t, http.MethodPut, completePath, nil)

		err := setup.Handler.GetTrustDomainByName(setup.EchoCtx, td1)
		assert.Error(t, err)

		echoHttpErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusNotFound, echoHttpErr.Code)
		assert.Equal(t, fmt.Sprintf("trust domain does not exist: %q", td1), echoHttpErr.Message)
	})
}

func TestUDSPutTrustDomainByName(t *testing.T) {
	trustDomainPath := "/trust-domain/%v"

	t.Run("Successfully updated a trust domain", func(t *testing.T) {
		fakeTrustDomains := entity.TrustDomain{ID: tdUUID1, Name: NewTrustDomain(t, td1)}

		completePath := fmt.Sprintf(trustDomainPath, tdUUID1.UUID)

		description := "I am being updated"
		reqBody := &admin.PutTrustDomainByNameJSONRequestBody{
			Id:          tdUUID1.UUID,
			Name:        td1,
			Description: &description,
		}

		// Setup
		setup := NewManagementTestSetup(t, http.MethodPut, completePath, reqBody)
		setup.FakeDatabase.WithTrustDomains(&fakeTrustDomains)

		err := setup.Handler.PutTrustDomainByName(setup.EchoCtx, td1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, setup.Recorder.Code)

		apiTrustDomain := api.TrustDomain{}
		err = json.Unmarshal(setup.Recorder.Body.Bytes(), &apiTrustDomain)
		assert.NoError(t, err)

		assert.Equal(t, td1, apiTrustDomain.Name)
		assert.Equal(t, tdUUID1.UUID, apiTrustDomain.Id)
		assert.Equal(t, description, *apiTrustDomain.Description)
	})

	t.Run("Raise a not found when trying to updated a trust domain that does not exists", func(t *testing.T) {
		completePath := fmt.Sprintf(trustDomainPath, tdUUID1.UUID)

		// Fake Request body
		description := "I am being updated"
		reqBody := &admin.PutTrustDomainByNameJSONRequestBody{
			Id:          tdUUID1.UUID,
			Name:        td1,
			Description: &description,
		}

		// Setup
		setup := NewManagementTestSetup(t, http.MethodPut, completePath, reqBody)

		err := setup.Handler.PutTrustDomainByName(setup.EchoCtx, td1)
		assert.Error(t, err)
		assert.Empty(t, setup.Recorder.Body.Bytes())

		echoHTTPErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusNotFound, echoHTTPErr.Code)
		expectedErrorMsg := fmt.Sprintf("trust domain does not exist: %q", td1)
		assert.Equal(t, expectedErrorMsg, echoHTTPErr.Message)
	})
}

func TestUDSGetJoinToken(t *testing.T) {
	trustDomainPath := "/trust-domain/%v/join-token"

	t.Run("Successfully generates a join token for the trust domain", func(t *testing.T) {
		td1ID := NewNullableID()
		fakeTrustDomains := entity.TrustDomain{ID: td1ID, Name: NewTrustDomain(t, td1)}

		completePath := fmt.Sprintf(trustDomainPath, td1)

		// Setup
		setup := NewManagementTestSetup(t, http.MethodGet, completePath, nil)
		setup.FakeDatabase.WithTrustDomains(&fakeTrustDomains)

		params := admin.GetJoinTokenParams{
			Ttl: 600,
		}
		err := setup.Handler.GetJoinToken(setup.EchoCtx, td1, params)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, setup.Recorder.Code)

		jtResp := admin.JoinTokenResponse{}
		err = json.Unmarshal(setup.Recorder.Body.Bytes(), &jtResp)
		assert.NoError(t, err)

		assert.NotEmpty(t, jtResp)
	})

	t.Run("Raise a bad request when trying to generates a join token for the trust domain that does not exists", func(t *testing.T) {
		completePath := fmt.Sprintf(trustDomainPath, td1)

		// Setup
		setup := NewManagementTestSetup(t, http.MethodGet, completePath, nil)

		params := admin.GetJoinTokenParams{
			Ttl: 600,
		}
		err := setup.Handler.GetJoinToken(setup.EchoCtx, td1, params)
		assert.Error(t, err)

		echoHttpErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, echoHttpErr.Code)

		expectedMsg := fmt.Sprintf("trust domain does not exist: %q", td1)
		assert.Equal(t, expectedMsg, echoHttpErr.Message)
	})
}

func NewNullableID() uuid.NullUUID {
	return uuid.NullUUID{
		Valid: true,
		UUID:  uuid.New(),
	}
}

func NewTrustDomain(t *testing.T, tdName string) spiffeid.TrustDomain {
	td, err := spiffeid.TrustDomainFromString(tdName)
	assert.NoError(t, err)
	return td
}

func assertContainRelationships(t *testing.T, expectedRelationships []*api.Relationship, actualRelationships []*api.Relationship) {
	require.Equal(t, len(expectedRelationships), len(actualRelationships))
	for _, expectedRel := range expectedRelationships {
		found := false
		for _, actualRel := range actualRelationships {
			if equalRelationships(expectedRel, actualRel) {
				found = true
				break
			}
		}
		require.True(t, found, "Did not find expected relationship in the actual relationships list: %+v", expectedRel)
	}
}

func equalRelationships(r1 *api.Relationship, r2 *api.Relationship) bool {
	return r1.Id == r2.Id && r1.TrustDomainAId == r2.TrustDomainAId && r1.TrustDomainBId == r2.TrustDomainBId &&
		r1.TrustDomainAConsent == r2.TrustDomainAConsent && r1.TrustDomainBConsent == r2.TrustDomainBConsent
}
