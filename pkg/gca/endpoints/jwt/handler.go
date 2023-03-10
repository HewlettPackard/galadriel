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

// this is the expected audience in the auth token and
// the audience that is set in the new tokens generated
// by this handler
const jwtAudience = "galadriel-ca"

type Handler struct {
	CA          ca.ServerCA
	Logger      logrus.FieldLogger
	JWTTokenTTL time.Duration
	Clock       clock.Clock
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debug("New JWT authToken requested")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is missing", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Invalid Authorization header format", http.StatusBadRequest)
		return
	}

	// Extract the JWT from the header
	bearerToken := strings.TrimPrefix(authHeader, "Bearer ")

	authToken, err := jwt.ParseSigned(bearerToken)
	if err != nil {
		http.Error(w, "Invalid JWT token", http.StatusBadRequest)
		return
	}

	claims := make(map[string]any)
	err = authToken.Claims(h.CA.GetPublicKey(), &claims)
	if err != nil {
		http.Error(w, "Error decoding JWT claims", http.StatusBadRequest)
		return
	}

	if h.isTokenExpired(claims) {
		http.Error(w, "Expired JWT token", http.StatusUnauthorized)
		return
	}

	aud, ok := claims["aud"].([]interface{})
	if !ok {
		http.Error(w, "Error decoding JWT audience", http.StatusBadRequest)
		return
	}

	if !containsAudience(aud) {
		http.Error(w, "Invalid JWT token audience", http.StatusUnauthorized)
		return
	}

	sub, err := spiffeid.FromString(claims["sub"].(string))
	if err != nil {
		http.Error(w, "Error parsing subject claim SPIFFE ID", http.StatusBadRequest)
		return
	}

	params := ca.JWTParams{
		Subject:  sub,
		Audience: []string{jwtAudience},
		TTL:      h.JWTTokenTTL,
	}

	newToken, err := h.CA.SignJWT(r.Context(), params)
	if err != nil {
		http.Error(w, "Error parsing subject field", http.StatusBadRequest)
		return
	}

	if _, err := w.Write([]byte(newToken)); err != nil {
		h.Logger.Errorf("Error writing token in HTTP response: %w", err)
	}
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

func containsAudience(aud []interface{}) bool {
	found := false
	for _, a := range aud {
		if a == jwtAudience {
			found = true
			break
		}
	}

	return found
}
