package endpoints

import (
	"encoding/json"
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

type ManagementTestSetup struct {
	EchoCtx      echo.Context
	Handler      *AdminAPIHandlers
	Recorder     *httptest.ResponseRecorder
	FakeDatabase *datastore.FakeDatabase
}

func NewManagementTestSetup(t *testing.T, method, url string, body interface{}) *ManagementTestSetup {
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
	fakeDB := datastore.NewFakeDB()
	logger := logrus.New()

	return &ManagementTestSetup{
		EchoCtx:      e.NewContext(req, rec),
		Recorder:     rec,
		Handler:      NewAdminAPIHandlers(logger, fakeDB),
		FakeDatabase: fakeDB,
	}
}

func TestUDSGetRelationships(t *testing.T) {
	t.Run("Successfully filter by trust domain", func(t *testing.T) {

		// Setup
		managementTestSetup := NewManagementTestSetup(t, http.MethodGet, "/relationships", nil)
		echoCtx := managementTestSetup.EchoCtx

		// Creating fake trust bundles and relationships to be filtered
		td1Name := NewTrustDomain(t, testTrustDomain)
		tdUUID1 := NewNullableID()
		tdUUID2 := NewNullableID()
		tdUUID3 := NewNullableID()

		fakeTrustDomains := []*entity.TrustDomain{
			{ID: tdUUID1, Name: td1Name},
			{ID: tdUUID2, Name: NewTrustDomain(t, "test2.com")},
			{ID: tdUUID3, Name: NewTrustDomain(t, "test3.com")},
		}

		fakeRelationships := []*entity.Relationship{
			{ID: NewNullableID(), TrustDomainAID: tdUUID1.UUID, TrustDomainBID: tdUUID2.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false},
			{ID: NewNullableID(), TrustDomainBID: tdUUID1.UUID, TrustDomainAID: tdUUID3.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false},
			{ID: NewNullableID(), TrustDomainAID: uuid.New(), TrustDomainBID: uuid.New(), TrustDomainAConsent: false, TrustDomainBConsent: false},
		}

		managementTestSetup.FakeDatabase.WithTrustDomains(fakeTrustDomains)
		managementTestSetup.FakeDatabase.WithRelationships(fakeRelationships)

		// managementTestSetup.Handler.Datastore
		tdName := td1Name.String()
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

		// apiRelations := mapRelationships([]*entity.Relationship{
		// 	{TrustDomainAID: tdUUID1.UUID, TrustDomainBID: tdUUID2.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false},
		// 	{TrustDomainBID: tdUUID1.UUID, TrustDomainAID: tdUUID3.UUID, TrustDomainAConsent: false, TrustDomainBConsent: false},
		// })

		// assert.ElementsMatchf(t, relationships, apiRelations, "filter does not work properly")
	})
}

func TestUDSPutRelationships(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestUDSGetRelationshipsRelationshipID(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestUDSPutTrustDomain(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestUDSGetTrustDomainTrustDomainName(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestUDSPutTrustDomainTrustDomainName(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
}

func TestUDSPostTrustDomainTrustDomainNameJoinToken(t *testing.T) {
	t.Skip("Missing tests will be added when the API be implemented")
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
