package jwt

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/ca"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmhodges/clock"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

const (
	// GCAAudience is the expected audience in the auth token and
	// the audience that is set in the new tokens generated
	// by this handler
	GCAAudience         = "galadriel-ca"
	GCAIssuer           = "galadriel-ca"
	AuthorizationHeader = "Authorization"
	Bearer              = "Bearer "
)

type Handler struct {
	serverCA    ca.ServerCA
	logger      logrus.FieldLogger
	jwtTokenTTL time.Duration
	clock       clock.Clock
}

type Config struct {
	ServerCA    ca.ServerCA
	Logger      logrus.FieldLogger
	JWTTokenTTL time.Duration
	Clock       clock.Clock
}

func NewHandler(c *Config) (http.Handler, error) {
	handler := &Handler{
		serverCA:    c.ServerCA,
		logger:      c.Logger,
		jwtTokenTTL: c.JWTTokenTTL,
		clock:       c.Clock,
	}

	return handleFunc(handler), nil
}

func handleFunc(handler *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler.logger.Debug("new JWT Token requested")

		if r.Method != http.MethodGet {
			http.Error(w, "method is not allowed", http.StatusMethodNotAllowed)
			return
		}

		jwtToken, ok := handler.getAuthJWTToken(w, r)
		if !ok {
			return
		}

		registeredClaims := jwtToken.Claims.(*jwt.RegisteredClaims)
		if ok := handler.validateClaims(w, registeredClaims); !ok {
			return
		}

		sub := registeredClaims.Subject
		// A valid SPIFFE trust domain name is expected
		subject, err := spiffeid.TrustDomainFromString(sub)
		if err != nil {
			http.Error(w, "invalid JWT token subject", http.StatusBadRequest)
			return
		}

		// params for the new JWT token
		params := ca.JWTParams{
			Issuer: GCAIssuer,
			// the new JWT token has the same subject as the received token
			Subject:  subject,
			Audience: []string{GCAAudience},
			TTL:      handler.jwtTokenTTL,
		}

		newToken, err := handler.serverCA.SignJWT(params)
		if err != nil {
			handler.logger.WithError(err).Error("Failed to generate JWT token")
			http.Error(w, "error generating new token", http.StatusInternalServerError)
			return
		}

		if _, err := w.Write([]byte(newToken)); err != nil {
			handler.logger.Errorf("error writing token in HTTP response: %w", err)
		}
	}
}

func (h *Handler) getAuthJWTToken(w http.ResponseWriter, r *http.Request) (*jwt.Token, bool) {
	authHeader := r.Header.Get(AuthorizationHeader)
	if strings.TrimSpace(authHeader) == "" {
		http.Error(w, "authorization header is missing", http.StatusBadRequest)
		return nil, false
	}
	if !strings.HasPrefix(authHeader, Bearer) {
		http.Error(w, "invalid authorization header format", http.StatusBadRequest)
		return nil, false
	}

	// Extract the JWT from the header
	bearerToken := strings.TrimPrefix(authHeader, Bearer)
	claims := &jwt.RegisteredClaims{}

	jwtToken, err := jwt.ParseWithClaims(bearerToken, claims, func(t *jwt.Token) (any, error) { return h.serverCA.PublicKey(), nil })
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			http.Error(w, "expired JWT token", http.StatusUnauthorized)
		} else {
			http.Error(w, "error decoding JWT claims", http.StatusBadRequest)
		}
		return nil, false
	}

	return jwtToken, true
}

func (h *Handler) validateClaims(w http.ResponseWriter, claims *jwt.RegisteredClaims) bool {
	if !containsAudience(claims.Audience, GCAAudience) {
		http.Error(w, "invalid JWT token audience", http.StatusUnauthorized)
		return false
	}

	return true
}

func containsAudience(aud []string, expected string) bool {
	for _, a := range aud {
		if a == expected {
			return true
		}
	}

	return false
}
