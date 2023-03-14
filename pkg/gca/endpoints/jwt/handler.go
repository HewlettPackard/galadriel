package jwt

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/ca"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/jmhodges/clock"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

const (
	// GCAAudience is the expected audience in the auth token and
	// the audience that is set in the new tokens generated
	// by this handler
	GCAAudience         = "galadriel-ca"
	AuthorizationHeader = "Authorization"
	Bearer              = "Bearer "
)

type Handler struct {
	CA          *ca.CA
	Logger      logrus.FieldLogger
	JWTTokenTTL time.Duration
	Clock       clock.Clock
}

type Config struct {
	CA          *ca.CA
	Logger      logrus.FieldLogger
	JWTTokenTTL time.Duration
	Clock       clock.Clock
}

func NewHandler(c *Config) (http.Handler, error) {
	handler := &Handler{
		CA:          c.CA,
		Logger:      c.Logger,
		JWTTokenTTL: c.JWTTokenTTL,
		Clock:       c.Clock,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.Logger.Debug("new JWT authToken requested")

		if r.Method != http.MethodGet {
			http.Error(w, "method is not allowed", http.StatusMethodNotAllowed)
			return
		}

		authHeader := r.Header.Get(AuthorizationHeader)
		if authHeader == "" {
			http.Error(w, "authorization header is missing", http.StatusBadRequest)
			return
		}
		if !strings.HasPrefix(authHeader, Bearer) {
			http.Error(w, "invalid authorization header format", http.StatusBadRequest)
			return
		}

		// Extract the JWT from the header
		bearerToken := strings.TrimPrefix(authHeader, Bearer)

		authToken, err := jwt.ParseSigned(bearerToken)
		if err != nil {
			http.Error(w, "invalid JWT token", http.StatusBadRequest)
			return
		}

		claims := make(map[string]any)
		err = authToken.Claims(handler.CA.PublicKey, &claims)
		if err != nil {
			http.Error(w, "error decoding JWT claims", http.StatusBadRequest)
			return
		}

		if handler.isTokenExpired(claims) {
			http.Error(w, "expired JWT token", http.StatusUnauthorized)
			return
		}

		aud, ok := claims["aud"].([]any)
		if !ok {
			http.Error(w, "error decoding JWT audience", http.StatusBadRequest)
			return
		}

		if !containsAudience(aud, GCAAudience) {
			http.Error(w, "invalid JWT token audience", http.StatusUnauthorized)
			return
		}

		sub := claims["sub"].(string)

		// A valid SPIFFE trust domain name is expected
		subject, err := spiffeid.TrustDomainFromString(sub)
		if err != nil {
			http.Error(w, "invalid JWT token subject", http.StatusBadRequest)
			return
		}

		// params for the new JWT token
		params := ca.JWTParams{
			Subject:  subject,
			Audience: []string{GCAAudience},
			TTL:      handler.JWTTokenTTL,
		}

		newToken, err := handler.CA.SignJWT(r.Context(), params)
		if err != nil {
			http.Error(w, "error generating new token", http.StatusBadRequest)
			return
		}

		if _, err := w.Write([]byte(newToken)); err != nil {
			handler.Logger.Errorf("error writing token in HTTP response: %w", err)
		}
	}), nil
}

func (h *Handler) isTokenExpired(claims map[string]any) bool {
	var expiration time.Time
	switch exp := claims["exp"].(type) {
	case float64:
		expiration = time.Unix(int64(exp), 0)
	case json.Number:
		v, _ := exp.Int64()
		expiration = time.Unix(v, 0)
	}

	return h.Clock.Now().After(expiration)
}

func containsAudience(aud []any, expected string) bool {
	for _, a := range aud {
		if a == expected {
			return true
		}
	}

	return false
}
