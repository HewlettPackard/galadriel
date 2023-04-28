package endpoints

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
)

type AuthNTestSetup struct {
	EchoCtx    echo.Context
	Middleware *AuthenticationMiddleware
	Recorder   *httptest.ResponseRecorder
	InMemoryDB *datastore.InMemoryDatabase
}

func SetupMiddleware() *AuthNTestSetup {
	logger := logrus.New()
	inMemoryDB := datastore.NewInMemoryDB()
	authnMiddleware := NewAuthenticationMiddleware(logger, inMemoryDB)

	e := echo.New()
	e.Use(middleware.KeyAuth(authnMiddleware.Authenticate))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	return &AuthNTestSetup{
		Recorder:   rec,
		InMemoryDB: inMemoryDB,
		Middleware: authnMiddleware,
		EchoCtx:    e.NewContext(req, rec),
	}
}

func SetupToken(t *testing.T, setup *AuthNTestSetup, token string) {
	td, err := spiffeid.TrustDomainFromString("test.com")
	assert.NoError(t, err)

	jt := &entity.JoinToken{
		Token:           token,
		TrustDomainID:   uuid.New(),
		TrustDomainName: td,
	}

	setup.Middleware.datastore.CreateJoinToken(context.TODO(), jt)
}

func TestAuthenticate(t *testing.T) {
	t.Run("Authorized tokens must be able to pass authn verification", func(t *testing.T) {
		authnSetup := SetupMiddleware()
		token := GenerateSecureToken(10)
		SetupToken(t, authnSetup, token)

		authorized, err := authnSetup.Middleware.Authenticate(token, authnSetup.EchoCtx)
		assert.NoError(t, err)
		assert.True(t, authorized)
	})

	t.Run("Problems when lookup data store must signalize internal server error", func(t *testing.T) {
		authnSetup := SetupMiddleware()
		authnSetup.InMemoryDB.FailNext()

		token := GenerateSecureToken(10)

		authorized, err := authnSetup.Middleware.Authenticate(token, authnSetup.EchoCtx)
		assert.NoError(t, err)
		assert.True(t, authorized)

		echoHTTPErr := err.(*echo.HTTPError)

		assert.Equal(t, http.StatusInternalServerError, echoHTTPErr.Code)
	})

	t.Run("Non authorized tokens must raise unauthorized responses", func(t *testing.T) {
		authnSetup := SetupMiddleware()

		token := GenerateSecureToken(10)

		authorized, err := authnSetup.Middleware.Authenticate(token, authnSetup.EchoCtx)
		assert.Error(t, err)
		assert.False(t, authorized)

		echoHTTPErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusUnauthorized, echoHTTPErr.Code)
	})
}

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
