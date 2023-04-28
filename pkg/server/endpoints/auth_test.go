package endpoints

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type AuthNTestSetup struct {
	EchoCtx    echo.Context
	Middleware *AuthenticationMiddleware
	Recorder   *httptest.ResponseRecorder
}

func SetupMiddleware() *AuthNTestSetup {
	logger := logrus.New()
	// ds := datastore.New()
	authnMiddleware := NewAuthenticationMiddleware(logger, _)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	return &AuthNTestSetup{
		Recorder:   rec,
		Middleware: authnMiddleware,
		EchoCtx:    e.NewContext(req, rec),
	}
}

func TestAuthenticate(t *testing.T) {
	t.Run("Authorized tokens must be able to pass authn verification", func(t *testing.T) {
		authnSetup := SetupMiddleware()
		token := ""

		authorized, err := authnSetup.Middleware.Authenticate(token, authnSetup.EchoCtx)
		assert.NoError(t, err)
		assert.True(t, authorized)
	})

	t.Run("Problems when lookup data store must signalize internal server error", func(t *testing.T) {
		authnSetup := SetupMiddleware()
		token := ""

		authorized, err := authnSetup.Middleware.Authenticate(token, authnSetup.EchoCtx)
		assert.NoError(t, err)
		assert.True(t, authorized)

		recorder := authnSetup.Recorder
		assert.Equal(t, recorder.Result().StatusCode, http.StatusInternalServerError)
	})

	t.Run("Non authorized tokens must raise unauthorized responses", func(t *testing.T) {
		authnSetup := SetupMiddleware()
		token := ""

		authorized, err := authnSetup.Middleware.Authenticate(token, authnSetup.EchoCtx)
		assert.Error(t, err)
		assert.False(t, authorized)

		recorder := authnSetup.Recorder
		assert.Equal(t, recorder.Result().StatusCode, http.StatusUnauthorized)
	})
}
