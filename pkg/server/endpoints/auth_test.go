package endpoints

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/common/jwt"
	"github.com/HewlettPackard/galadriel/pkg/common/keymanager"
	"github.com/HewlettPackard/galadriel/test/fakes/fakedatastore"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type AuthNTestSetup struct {
	EchoCtx      echo.Context
	Middleware   *AuthenticationMiddleware
	Recorder     *httptest.ResponseRecorder
	FakeDatabase *fakedatastore.FakeDatabase
	JWTIssuer    jwt.Issuer
}

func SetupMiddleware(t *testing.T) *AuthNTestSetup {
	logger := logrus.New()
	fakeDB := fakedatastore.NewFakeDB()

	km := keymanager.NewMemoryKeyManager(nil)
	c := jwt.ValidatorConfig{
		KeyManager:       km,
		ExpectedAudience: []string{"test"},
	}
	jwtValidator := jwt.NewDefaultJWTValidator(&c)

	key, err := km.GenerateKey(context.Background(), "test-key-id", cryptoutil.RSA2048)
	require.NoError(t, err)

	jwtIssuer, err := jwt.NewJWTCA(&jwt.Config{
		Signer: key.Signer(),
		Kid:    "test-key-id",
	})
	require.NoError(t, err)

	authnMiddleware := NewAuthenticationMiddleware(logger, fakeDB, jwtValidator)

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
		JWTIssuer:    jwtIssuer,
	}
}

func TestAuthenticate(t *testing.T) {
	t.Run("Authorized tokens must be able to pass authn verification", func(t *testing.T) {
		authnSetup := SetupMiddleware(t)

		td := entity.TrustDomain{Name: spiffeid.RequireTrustDomainFromString("spiffe://test.com")}
		authnSetup.FakeDatabase.WithTrustDomains(&td)

		token, err := authnSetup.JWTIssuer.IssueJWT(context.Background(), &jwt.JWTParams{
			Issuer:   "test",
			Subject:  td.Name,
			Audience: []string{"test"},
			TTL:      5 * time.Minute,
		})
		require.NoError(t, err)

		authorized, err := authnSetup.Middleware.Authenticate(token, authnSetup.EchoCtx)
		assert.NoError(t, err)
		assert.True(t, authorized)
	})

	t.Run("Non authorized tokens must raise unauthorized responses", func(t *testing.T) {
		authnSetup := SetupMiddleware(t)

		token, err := authnSetup.JWTIssuer.IssueJWT(context.Background(), &jwt.JWTParams{
			Issuer:   "test",
			Subject:  spiffeid.RequireTrustDomainFromString("spiffe://test.com/test"),
			Audience: []string{"invalid-audience"},
			TTL:      5 * time.Minute,
		})
		require.NoError(t, err)

		authorized, err := authnSetup.Middleware.Authenticate(token, authnSetup.EchoCtx)
		assert.Error(t, err)
		assert.False(t, authorized)

		echoHTTPErr := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusUnauthorized, echoHTTPErr.Code)
	})
}
