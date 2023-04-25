package endpoints

import (
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type AuthenthicationMD struct {
	Datastore datastore.Datastore
	Logger    logrus.FieldLogger
}

func (m AuthenthicationMD) Authenticate(token string, ctx echo.Context) (bool, error) {

	gctx := ctx.Request().Context()

	// Any skip cases ?
	t, err := m.Datastore.FindJoinToken(gctx, token)
	if err != nil {
		m.Logger.Errorf("Invalid Token: %s\n", token)
		return false, err
	}

	m.Logger.Debugf("Token valid for trust domain: %s\n", t.TrustDomainID)

	ctx.Set("token", t)

	return true, nil
}
