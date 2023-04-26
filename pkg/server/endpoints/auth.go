package endpoints

import (
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type AuthenticationMiddleware struct {
	datastore datastore.Datastore
	logger    logrus.FieldLogger
}

func NewAuthenticationMiddleware(l logrus.FieldLogger, ds datastore.Datastore) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		logger:    l,
		datastore: ds,
	}
}

func (m AuthenticationMiddleware) Authenticate(token string, echoCtx echo.Context) (bool, error) {
	ctx := echoCtx.Request().Context()

	// Any skip cases ?
	t, err := m.datastore.FindJoinToken(ctx, token)
	if err != nil {
		message := "Invalid authorization token"
		return false, echo.NewHTTPError(http.StatusUnauthorized, message)
	}

	if t == nil {
		message := "Token not found"
		return false, echo.NewHTTPError(http.StatusUnauthorized, message)
	}

	m.logger.Debugf("Token valid for trust domain: %s\n", t.TrustDomainID)
	echoCtx.Set("token", t)

	return true, nil
}
