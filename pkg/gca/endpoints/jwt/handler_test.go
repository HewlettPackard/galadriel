package jwt

import (
	"crypto"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/ca"
	"github.com/HewlettPackard/galadriel/test/certtest"
	"github.com/HewlettPackard/galadriel/test/jwttest"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/jmhodges/clock"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	logger, _ := test.NewNullLogger()
	clk := clock.NewFake()

	CA, signer := createCA(t)
	handler, err := NewHandler(&Config{
		CA:          CA,
		Logger:      logger,
		JWTTokenTTL: 100,
		Clock:       clk,
	})
	require.NoError(t, err)

	kid := jwttest.GenerateRandomKeyID(t)

	server := httptest.NewServer(handler)
	defer server.Close()

	testCases := []struct {
		name         string
		errorMessage string
		statusCode   int
		call         func(server *httptest.Server) (*http.Response, error)
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			call: func(server *httptest.Server) (*http.Response, error) {
				token := jwttest.CreateToken(t, clk, signer, kid, "domain.test", []string{GCAAudience}, time.Hour)
				req := buildRequest(t, CA, server.URL, token)
				resp := doRequest(t, req)
				return resp, nil
			},
		},
		{
			name:         "method not allows",
			statusCode:   http.StatusMethodNotAllowed,
			errorMessage: "method is not allowed\n",
			call: func(server *httptest.Server) (*http.Response, error) {
				req, err := http.NewRequest("POST", server.URL, nil)
				require.NoError(t, err)
				resp := doRequest(t, req)
				return resp, nil
			},
		},
		{
			name:         "missing authorization header",
			statusCode:   http.StatusBadRequest,
			errorMessage: "authorization header is missing\n",
			call: func(server *httptest.Server) (*http.Response, error) {
				req, err := http.NewRequest("GET", server.URL, nil)
				require.NoError(t, err)
				resp := doRequest(t, req)
				return resp, nil
			},
		},
		{
			name:         "invalid authorization header",
			statusCode:   http.StatusBadRequest,
			errorMessage: "invalid authorization header format\n",
			call: func(server *httptest.Server) (*http.Response, error) {
				req, err := http.NewRequest("GET", server.URL, nil)
				require.NoError(t, err)

				token := jwttest.CreateToken(t, clk, signer, kid, "domain.test", []string{GCAAudience}, time.Hour)
				req.Header.Set("Authorization", fmt.Sprintf("Wrong %s", token))

				resp := doRequest(t, req)
				return resp, nil
			},
		},
		{
			name:         "invalid token",
			statusCode:   http.StatusBadRequest,
			errorMessage: "invalid JWT token\n",
			call: func(server *httptest.Server) (*http.Response, error) {
				req, err := http.NewRequest("GET", server.URL, nil)
				require.NoError(t, err)

				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "not-a-token"))

				resp := doRequest(t, req)
				return resp, nil
			},
		},
		{
			name:         "expired token",
			statusCode:   http.StatusUnauthorized,
			errorMessage: "expired JWT token\n",
			call: func(server *httptest.Server) (*http.Response, error) {
				req, err := http.NewRequest("GET", server.URL, nil)
				require.NoError(t, err)

				minuteAgo := -1 * time.Minute
				token := jwttest.CreateToken(t, clk, signer, kid, "domain.test", []string{GCAAudience}, minuteAgo)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

				resp := doRequest(t, req)
				return resp, nil
			},
		},
		{
			name:         "invalid token audience",
			statusCode:   http.StatusUnauthorized,
			errorMessage: "invalid JWT token audience\n",
			call: func(server *httptest.Server) (*http.Response, error) {
				req, err := http.NewRequest("GET", server.URL, nil)
				require.NoError(t, err)

				token := jwttest.CreateToken(t, clk, signer, kid, "domain.test", []string{"other-audience"}, time.Hour)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

				resp := doRequest(t, req)
				return resp, nil
			},
		},
		{
			name:         "invalid token subject",
			statusCode:   http.StatusBadRequest,
			errorMessage: "invalid JWT token subject\n",
			call: func(server *httptest.Server) (*http.Response, error) {
				req, err := http.NewRequest("GET", server.URL, nil)
				require.NoError(t, err)

				token := jwttest.CreateToken(t, clk, signer, kid, "unix:/not-a-domain-name", []string{GCAAudience}, time.Hour)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

				resp := doRequest(t, req)
				return resp, nil
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			res, err := testCase.call(server)
			require.NoError(t, err)
			defer res.Body.Close()

			actual, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			switch {
			case res.StatusCode == http.StatusOK:
				_, err = jwt.ParseSigned(string(actual))
				require.NoError(t, err)
			default:
				require.Equal(t, testCase.statusCode, res.StatusCode)
				require.Equal(t, testCase.errorMessage, string(actual))
			}
		})
	}
}

func doRequest(t *testing.T, req *http.Request) *http.Response {
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func buildRequest(t *testing.T, CA *ca.CA, serverURL string, token string) *http.Request {
	req, err := http.NewRequest("GET", serverURL, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return req
}

func createCA(t *testing.T) (*ca.CA, crypto.Signer) {
	clk := clock.NewFake()
	caCert, caKey, err := certtest.CreateTestCACertificate(clk)
	require.NoError(t, err)

	caConfig := &ca.Config{
		RootCert: caCert,
		RootKey:  caKey,
		Clock:    clk,
	}
	CA, err := ca.New(caConfig)
	require.NoError(t, err)

	return CA, caKey.(crypto.Signer)
}
