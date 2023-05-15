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
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
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

	// Trust Domains ID's
	tdUUID1 = NewNullableID()
	tdUUID2 = NewNullableID()
	tdUUID3 = NewNullableID()
)

type ManagementTestSetup struct {
	EchoCtx      echo.Context
	Handler      *AdminAPIHandlers
	Recorder     *httptest.ResponseRecorder
	FakeDatabase *datastore.FakeDatabase

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
	fakeDB := datastore.NewFakeDB()
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

func TestUDSGetRelationships(t *testing.T) {
	relationshipsPath := "/relationships"

	t.Run("Successfully filter by trust domain", func(t *testing.T) {
		// Setup
		managementTestSetup := NewManagementTestSetup(t, http.MethodGet, relationshipsPath, nil)
		echoCtx := managementTestSetup.EchoCtx

		td1Name := NewTrustDomain(t, td1)

		fakeTrustDomains := []*entity.TrustDomain{
			{ID: tdUUID1, Name: td1Name},
			{ID: tdUUID2, Name: NewTrustDomain(t, td2)},
			{ID: tdUUID3, Name: NewTrustDomain(t, td3)},
		}

		fakeRelationships := []*entity.Relationship{
			{ID: r1ID, TrustDomainAID: tdUUID1.UUID, TrustDomainBID: tdUUID2.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false},
			{ID: r2ID, TrustDomainBID: tdUUID1.UUID, TrustDomainAID: tdUUID3.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false},
			{ID: NewNullableID(), TrustDomainAID: uuid.New(), TrustDomainBID: uuid.New(), TrustDomainAConsent: true, TrustDomainBConsent: true},
			{ID: NewNullableID(), TrustDomainAID: uuid.New(), TrustDomainBID: uuid.New(), TrustDomainAConsent: true, TrustDomainBConsent: false},
			{ID: NewNullableID(), TrustDomainAID: uuid.New(), TrustDomainBID: uuid.New(), TrustDomainAConsent: false, TrustDomainBConsent: false},
		}

		managementTestSetup.FakeDatabase.WithTrustDomains(fakeTrustDomains...)
		managementTestSetup.FakeDatabase.WithRelationships(fakeRelationships...)

		tdName := td1
		params := admin.GetRelationshipsParams{
			TrustDomainName: &tdName,
		}

		err := managementTestSetup.Handler.GetRelationships(echoCtx, params)
		assert.NoError(t, err)

		recorder := managementTestSetup.Recorder
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.NotEmpty(t, recorder.Body)

		relationships := []*api.Relationship{}
		err = json.Unmarshal(recorder.Body.Bytes(), &relationships)
		assert.NoError(t, err)

		assert.Len(t, relationships, 2)

		apiRelations := mapRelationships([]*entity.Relationship{
			{ID: r1ID, TrustDomainAID: tdUUID1.UUID, TrustDomainBID: tdUUID2.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false},
			{ID: r2ID, TrustDomainBID: tdUUID1.UUID, TrustDomainAID: tdUUID3.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false},
		})

		assert.ElementsMatch(t, relationships, apiRelations, "trust domain name filter does not work properly")
	})

	t.Run("Successfully filter by status", func(t *testing.T) {
		// Setup
		setup := NewManagementTestSetup(t, http.MethodGet, relationshipsPath, nil)

		td1Name := NewTrustDomain(t, td1)

		fakeTrustDomains := []*entity.TrustDomain{
			{ID: tdUUID1, Name: td1Name},
			{ID: tdUUID2, Name: NewTrustDomain(t, td2)},
			{ID: tdUUID3, Name: NewTrustDomain(t, td3)},
		}

		fakeRelationships := []*entity.Relationship{
			{ID: r1ID, TrustDomainAID: tdUUID1.UUID, TrustDomainBID: tdUUID3.UUID, TrustDomainAConsent: true, TrustDomainBConsent: true},
			{ID: r2ID, TrustDomainAID: tdUUID1.UUID, TrustDomainBID: tdUUID2.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false},
			{ID: r3ID, TrustDomainAID: tdUUID2.UUID, TrustDomainBID: tdUUID3.UUID, TrustDomainAConsent: false, TrustDomainBConsent: true},
		}

		setup.FakeDatabase.WithTrustDomains(fakeTrustDomains...)
		setup.FakeDatabase.WithRelationships(fakeRelationships...)

		expectedRelations := mapRelationships([]*entity.Relationship{
			{ID: r1ID, TrustDomainAID: tdUUID1.UUID, TrustDomainBID: tdUUID3.UUID, TrustDomainAConsent: true, TrustDomainBConsent: true},
		})

		assertFilter(t, setup, expectedRelations, admin.Approved)

		expectedRelations = mapRelationships([]*entity.Relationship{
			{ID: r2ID, TrustDomainAID: tdUUID1.UUID, TrustDomainBID: tdUUID2.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false},
			{ID: r3ID, TrustDomainAID: tdUUID2.UUID, TrustDomainBID: tdUUID3.UUID, TrustDomainAConsent: false, TrustDomainBConsent: true},
		})

		assertFilter(t, setup, expectedRelations, admin.Denied)

		expectedRelations = mapRelationships([]*entity.Relationship{
			{ID: r2ID, TrustDomainAID: tdUUID1.UUID, TrustDomainBID: tdUUID2.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false},
			{ID: r3ID, TrustDomainAID: tdUUID2.UUID, TrustDomainBID: tdUUID3.UUID, TrustDomainAConsent: false, TrustDomainBConsent: true},
		})

		assertFilter(t, setup, expectedRelations, admin.Pending)
	})

	t.Run("Should raise a bad request when receiving undefined status filter", func(t *testing.T) {

		// Setup
		setup := NewManagementTestSetup(t, http.MethodGet, relationshipsPath, nil)

		// Approved filter
		var randomFilter admin.GetRelationshipsParamsStatus = "a random filter"
		params := admin.GetRelationshipsParams{
			Status: &randomFilter,
		}

		err := setup.Handler.GetRelationships(setup.EchoCtx, params)
		assert.Error(t, err)

		httpErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Empty(t, setup.Recorder.Body)

		expectedMsg := fmt.Sprintf(
			"unrecognized status filter %v, accepted values [%v, %v, %v]",
			randomFilter, admin.Denied, admin.Approved, admin.Pending,
		)

		assert.ErrorContains(t, err, expectedMsg)
	})
}

func assertFilter(
	t *testing.T,
	setup *ManagementTestSetup,
	expectedRelations []*api.Relationship,
	status admin.GetRelationshipsParamsStatus,
) {
	setup.Refresh()

	strAddress := status
	params := admin.GetRelationshipsParams{
		Status: &strAddress,
	}

	err := setup.Handler.GetRelationships(setup.EchoCtx, params)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, setup.Recorder.Code)
	assert.NotEmpty(t, setup.Recorder.Body)

	relationships := []*api.Relationship{}
	err = json.Unmarshal(setup.Recorder.Body.Bytes(), &relationships)
	assert.NoError(t, err)

	assert.Len(t, relationships, len(expectedRelations))

	assert.ElementsMatchf(t, relationships, expectedRelations, "%v status filter does not work properly", status)
}

func TestUDSPutRelationships(t *testing.T) {
	relationshipsPath := "/relationships"

	t.Run("Successfully create a new relationship request", func(t *testing.T) {

		fakeTrustDomains := []*entity.TrustDomain{
			{ID: tdUUID1, Name: NewTrustDomain(t, td1)},
			{ID: tdUUID2, Name: NewTrustDomain(t, td2)},
		}

		reqBody := &admin.PutRelationshipJSONRequestBody{
			TrustDomainAId: tdUUID1.UUID,
			TrustDomainBId: tdUUID2.UUID,
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
			TrustDomainAId: tdUUID1.UUID,
			TrustDomainBId: tdUUID2.UUID,
		}

		// Setup
		setup := NewManagementTestSetup(t, http.MethodPut, relationshipsPath, reqBody)
		setup.FakeDatabase.WithTrustDomains(fakeTrustDomains...)

		err := setup.Handler.PutRelationship(setup.EchoCtx)
		assert.Error(t, err)
		assert.Empty(t, setup.Recorder.Body.Bytes())

		echoHTTPErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, echoHTTPErr.Code)

		expectedErrorMsg := fmt.Sprintf("trust domain %v does not exists", tdUUID2.UUID)
		assert.Equal(t, expectedErrorMsg, echoHTTPErr.Message)
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
		reqBody := &admin.TrustDomainPut{
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
		reqBody := &admin.TrustDomainPut{
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
		assert.ErrorContains(t, echoHttpErr, "trust domain already exists")
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

	t.Run("Raise a not found when trying to retrieve a trust domain that does not exists", func(t *testing.T) {
		completePath := fmt.Sprintf(trustDomainPath, tdUUID1.UUID)

		// Setup
		setup := NewManagementTestSetup(t, http.MethodPut, completePath, nil)

		err := setup.Handler.GetTrustDomainByName(setup.EchoCtx, td1)
		assert.Error(t, err)

		echoHttpErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusNotFound, echoHttpErr.Code)
		assert.Equal(t, "trust domain does not exists", echoHttpErr.Message)
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

		err := setup.Handler.PutTrustDomainByName(setup.EchoCtx, tdUUID1.UUID)
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

		err := setup.Handler.PutTrustDomainByName(setup.EchoCtx, tdUUID1.UUID)
		assert.Error(t, err)
		assert.Empty(t, setup.Recorder.Body.Bytes())

		echoHTTPErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusNotFound, echoHTTPErr.Code)
		expectedErrorMsg := fmt.Sprintf("trust domain %v does not exists", tdUUID1.UUID)
		assert.Equal(t, expectedErrorMsg, echoHTTPErr.Message)
	})
}

// func TestUDSPostTrustDomainTrustDomainNameJoinToken(t *testing.T) {
// 	trustDomainPath := "/trust-domain/%v/join-token"

// 	t.Run("Successfully generates a join token for the trust domain", func(t *testing.T) {
// 		td1ID := NewNullableID()
// 		fakeTrustDomains := entity.TrustDomain{ID: td1ID, Name: NewTrustDomain(t, td1)}

// 		completePath := fmt.Sprintf(trustDomainPath, td1)

// 		// Setup
// 		setup := NewManagementTestSetup(t, http.MethodPut, completePath, nil)
// 		setup.FakeDatabase.WithTrustDomains(&fakeTrustDomains)

// 		err := setup.Handler.PostTrustDomainTrustDomainNameJoinToken(setup.EchoCtx, td1)
// 		assert.NoError(t, err)
// 		assert.Equal(t, http.StatusOK, setup.Recorder.Code)

// 		apiJToken := admin.JoinTokenResult{}
// 		err = json.Unmarshal(setup.Recorder.Body.Bytes(), &apiJToken)
// 		assert.NoError(t, err)

// 		assert.NotEmpty(t, apiJToken)
// 	})

// 	t.Run("Raise a bad request when trying to generates a join token for the trust domain that does not exists", func(t *testing.T) {
// 		completePath := fmt.Sprintf(trustDomainPath, td1)

// 		// Setup
// 		setup := NewManagementTestSetup(t, http.MethodPut, completePath, nil)

// 		err := setup.Handler.PostTrustDomainTrustDomainNameJoinToken(setup.EchoCtx, td1)
// 		assert.Error(t, err)

// 		echoHttpErr := err.(*echo.HTTPError)
// 		assert.Equal(t, http.StatusBadRequest, echoHttpErr.Code)

// 		expectedMsg := fmt.Sprintf("trust domain %v does not exists ", td1)
// 		assert.Equal(t, expectedMsg, echoHttpErr.Message)
// 	})
// }

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
