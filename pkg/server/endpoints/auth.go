package endpoints

import (
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common/jwt"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type AuthenticationMiddleware struct {
	datastore    datastore.Datastore
	jwtValidator jwt.Validator
	logger       logrus.FieldLogger
}

func NewAuthenticationMiddleware(l logrus.FieldLogger, ds datastore.Datastore, jwtValidator jwt.Validator) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		logger:       l,
		datastore:    ds,
		jwtValidator: jwtValidator,
	}
}

// Authenticate is the middleware method that is responsible for authenticating the calling Harvester using the JWT token
// passed in the Authorization header.
func (m *AuthenticationMiddleware) Authenticate(bearerToken string, echoCtx echo.Context) (bool, error) {
	ctx := echoCtx.Request().Context()

	claims, err := m.jwtValidator.ValidateToken(ctx, bearerToken)
	if err != nil {
		return false, echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	subject := claims.Subject
	if subject == "" {
		return false, echo.NewHTTPError(http.StatusUnauthorized, "invalid token: missing subject")
	}

	tdName, err := spiffeid.TrustDomainFromString(subject)
	if err != nil {
		return false, echo.NewHTTPError(http.StatusUnauthorized, "invalid token: invalid trust domain name")
	}

	td, err := m.datastore.FindTrustDomainByName(ctx, tdName)
	if err != nil {
		return false, echo.NewHTTPError(http.StatusUnauthorized, "invalid token: trust domain not found")
	}

	m.logger.Debugf("Token valid for trust domain: %s\n", tdName)

	// set the authenticated trust domain ID in the echo context
	echoCtx.Set(authTrustDomainKey, td)
	// set the authenticated claims in the echo context
	echoCtx.Set(authClaimsKey, claims)

	return true, nil
}
