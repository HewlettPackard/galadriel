package endpoints

import (
	"fmt"
	"net/http"

	chttp "github.com/HewlettPackard/galadriel/pkg/common/http"
	"github.com/HewlettPackard/galadriel/pkg/common/jwt"
	"github.com/HewlettPackard/galadriel/pkg/server/db"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type AuthenticationMiddleware struct {
	datastore    db.Datastore
	jwtValidator jwt.Validator
	logger       logrus.FieldLogger
}

func NewAuthenticationMiddleware(l logrus.FieldLogger, ds db.Datastore, jwtValidator jwt.Validator) *AuthenticationMiddleware {
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
		msg := "invalid JWT authentication token"
		return false, chttp.LogAndRespondWithError(m.logger, err, msg, http.StatusUnauthorized)
	}

	subject := claims.Subject
	if subject == "" {
		return false, chttp.LogAndRespondWithError(m.logger, err, "invalid token: missing subject", http.StatusUnauthorized)
	}

	tdName, err := spiffeid.TrustDomainFromString(subject)
	if err != nil {
		return false, chttp.LogAndRespondWithError(m.logger, err, "invalid token: invalid trust domain name", http.StatusUnauthorized)
	}

	td, err := m.datastore.FindTrustDomainByName(ctx, tdName)
	if err != nil {
		return false, chttp.LogAndRespondWithError(m.logger, err, "invalid token: trust domain not found", http.StatusUnauthorized)
	}

	if td == nil {
		msg := fmt.Sprintf("trust domain not found: %q", tdName)
		return false, chttp.LogAndRespondWithError(m.logger, nil, msg, http.StatusUnauthorized)
	}

	// set the authenticated trust domain ID in the echo context
	echoCtx.Set(authTrustDomainKey, td)
	// set the authenticated claims in the echo context
	echoCtx.Set(authClaimsKey, claims)

	return true, nil
}
