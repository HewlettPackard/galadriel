package endpoints

import (
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type AuthenthicationMD struct {
	Datastore datastore.Datastore
	Logger    logrus.FieldLogger
}

func NewAuthenthicationMiddleware(l logrus.FieldLogger, ds datastore.Datastore) AuthenthicationMD {
	return AuthenthicationMD{
		Logger:    l,
		Datastore: ds,
	}
}

func (m AuthenthicationMD) Authenticate(token string, ctx echo.Context) (bool, error) {
	gctx := ctx.Request().Context()

	// Any skip cases ?
	t, err := m.Datastore.FindJoinToken(gctx, token)
	if err != nil {
		m.Logger.Errorf("Invalid Token: %s\n", token)
		message := "Invalid authorization token"
		return false, echo.NewHTTPError(http.StatusUnauthorized, message)
	}

	if t == nil {
		message := "Token not found"
		return false, echo.NewHTTPError(http.StatusForbidden, message)
	}

	m.Logger.Debugf("Token valid for trust domain: %s\n", t.TrustDomainID)
	ctx.Set("token", t)

	return true, nil
}
