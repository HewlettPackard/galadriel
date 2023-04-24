package endpoints

import (
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type AuthNMiddleware struct {
	Datastore datastore.Datastore
	Logger    logrus.FieldLogger
}

func (m AuthNMiddleware) AuthNF(token string, ctx echo.Context) (bool, error) {
	// Any skip cases ?
	t, err := m.Datastore.FindJoinToken(ctx.Request().Context(), token)
	if err != nil {
		m.Logger.Errorf("Invalid Token: %s\n", token)
		return false, err
	}

	m.Logger.Debugf("Token valid for trust domain: %s\n", t.TrustDomainID)

	ctx.Set("token", t)

	return true, nil
}
