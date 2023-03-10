package jwt

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/ca"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/jmhodges/clock"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeHTTP(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	clk := clock.NewFake()
	caConfig := &ca.Config{
		RootCertFile: "../testdata/root_cert.pem",
		RootKeyFile:  "../testdata/root_key.pem",
		Clock:        clk,
	}
	CA, err := ca.New(caConfig)
	require.NoError(t, err)

	params := ca.JWTParams{
		Subject:  spiffeid.RequireFromString("spiffe://example/test"),
		Audience: []string{jwtAudience},
		TTL:      time.Hour,
	}
	token, err := CA.SignJWT(context.Background(), params)
	require.NoError(t, err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	logger, _ := test.NewNullLogger()
	h := Handler{
		CA:          CA,
		Logger:      logger,
		JWTTokenTTL: time.Hour,
		Clock:       clk,
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.ServeHTTP)

	// Call ServeHTTP
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	response := rr.Body.String()
	parsed, err := jwt.ParseSigned(response)
	require.NoError(t, err)
	require.NotNil(t, parsed)

	publicKey := CA.PublicKey

	claims := make(map[string]any)
	err = parsed.Claims(publicKey, &claims)
	require.NoError(t, err)
	assert.Equal(t, "spiffe://example/test", claims["sub"])
	assert.Equal(t, float64(3600), claims["exp"])
	assert.Contains(t, claims["aud"], jwtAudience)
}

func TestServeHTTPExpiredToken(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	clk := clock.NewFake()
	caConfig := &ca.Config{
		RootCertFile: "../testdata/root_cert.pem",
		RootKeyFile:  "../testdata/root_key.pem",
		Clock:        clk,
	}
	CA, err := ca.New(caConfig)
	require.NoError(t, err)

	params := ca.JWTParams{
		Subject:  spiffeid.RequireFromString("spiffe://example/test"),
		Audience: []string{jwtAudience},
		TTL:      -1,
	}
	token, err := CA.SignJWT(context.Background(), params)
	require.NoError(t, err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	logger, _ := test.NewNullLogger()
	h := Handler{
		CA:          CA,
		Logger:      logger,
		JWTTokenTTL: time.Hour,
		Clock:       clk,
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.ServeHTTP)

	// Call ServeHTTP
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	message := strings.Replace(rr.Body.String(), "\n", "", -1)
	assert.Equal(t, "Expired JWT token", message)
}

func TestServeHTTPInvalidAudience(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	clk := clock.NewFake()
	caConfig := &ca.Config{
		RootCertFile: "../testdata/root_cert.pem",
		RootKeyFile:  "../testdata/root_key.pem",
		Clock:        clk,
	}
	CA, err := ca.New(caConfig)
	require.NoError(t, err)

	params := ca.JWTParams{
		Subject:  spiffeid.RequireFromString("spiffe://example/test"),
		Audience: []string{"other-audience"},
		TTL:      time.Hour,
	}
	token, err := CA.SignJWT(context.Background(), params)
	require.NoError(t, err)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	logger, _ := test.NewNullLogger()
	h := Handler{
		CA:          CA,
		Logger:      logger,
		JWTTokenTTL: time.Hour,
		Clock:       clk,
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.ServeHTTP)

	// Call ServeHTTP
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	message := strings.Replace(rr.Body.String(), "\n", "", -1)
	assert.Equal(t, "Invalid JWT token audience", message)
}
