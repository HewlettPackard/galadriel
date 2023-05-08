package endpoints

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type AuthNTestSetup struct {
	EchoCtx      echo.Context
	Middleware   *AuthenticationMiddleware
	Recorder     *httptest.ResponseRecorder
	FakeDatabase *datastore.FakeDatabase
}

func SetupMiddleware() *AuthNTestSetup {
	logger := logrus.New()
	fakeDB := datastore.NewFakeDB()
	authnMiddleware := NewAuthenticationMiddleware(logger, fakeDB)

	e := echo.New()
	e.Use(middleware.KeyAuth(authnMiddleware.Authenticate))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	return &AuthNTestSetup{
		Recorder:     rec,
		FakeDatabase: fakeDB,
		Middleware:   authnMiddleware,
		EchoCtx:      e.NewContext(req, rec),
	}
}

func setupToken(t *testing.T, echoCtx echo.Context, ds datastore.Datastore, td *entity.TrustDomain) *entity.JoinToken {
	jt := &entity.JoinToken{
		Token:           uuid.NewString(),
		TrustDomainID:   td.ID.UUID,
		TrustDomainName: td.Name,
	}

	jt, err := ds.CreateJoinToken(context.Background(), jt)
	require.NoError(t, err)
	require.NotNil(t, jt)

	echoCtx.Set(tokenKey, jt)

	return jt
}

func TestAuthenticate(t *testing.T) {
	t.Run("Authorized tokens must be able to pass authn verification", func(t *testing.T) {
		authnSetup := SetupMiddleware()
		token := setupToken(t, authnSetup.EchoCtx, authnSetup.FakeDatabase, testTrustDomain)

		authorized, err := authnSetup.Middleware.Authenticate(token.Token, authnSetup.EchoCtx)

		assert.NoError(t, err)
		assert.True(t, authorized)
	})

	t.Run("Problems when lookup data store must signalize internal server error", func(t *testing.T) {
		authnSetup := SetupMiddleware()

		expectedErr := errors.New("connection error")
		authnSetup.FakeDatabase.SetNextError(expectedErr)

		token := uuid.NewString()

		authorized, err := authnSetup.Middleware.Authenticate(token, authnSetup.EchoCtx)
		assert.Error(t, err)
		assert.False(t, authorized)

		echoHTTPErr := err.(*echo.HTTPError)
		assert.Equal(t, expectedErr.Error(), echoHTTPErr.Message)
		assert.Equal(t, http.StatusInternalServerError, echoHTTPErr.Code)
	})

	t.Run("Non authorized tokens must raise unauthorized responses", func(t *testing.T) {
		authnSetup := SetupMiddleware()

		token := uuid.NewString()

		authorized, err := authnSetup.Middleware.Authenticate(token, authnSetup.EchoCtx)
		assert.Error(t, err)
		assert.False(t, authorized)

		echoHTTPErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusUnauthorized, echoHTTPErr.Code)
	})
}

// func generateFakeToken(length int) string {

// 	 uuid.NewUUID()
// 	b := make([]byte, length)
// 	if _, err := rand.Read(b); err != nil {
// 		return ""
// 	}
// 	return hex.EncodeToString(b)
// }
