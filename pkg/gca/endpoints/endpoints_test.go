package endpoints

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/ca"
	"github.com/HewlettPackard/galadriel/test/certtest"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmhodges/clock"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	config, _ := newEndpointTestConfig(t)

	endpoint, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, endpoint)

	assert.Equal(t, endpoint.serverCA, config.ServerCA)
	assert.Equal(t, endpoint.tcpAddress, config.TCPAddress)
	assert.Equal(t, endpoint.localAddr, config.LocalAddress)
	assert.Equal(t, endpoint.logger, config.Logger)
}

func TestListenAndServe(t *testing.T) {
	config, caCert := newEndpointTestConfig(t)

	endpoints, err := New(config)
	require.NoError(t, err)

	endpoints.hooks.tcpListening = make(chan struct{})

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	errCh := make(chan error)
	go func() {
		errCh <- endpoints.ListenAndServe(ctx)
	}()
	defer func() {
		cancel()
		assert.NoError(t, <-errCh)
	}()

	waitForListening(t, endpoints, errCh)

	// create RootCertPool for TLS Client using the caCert from the GCA endpoints
	rootCa := x509.NewCertPool()
	require.NoError(t, err)
	rootCa.AddCert(caCert)

	tlsConfig := &tls.Config{
		RootCAs:    rootCa,
		ServerName: serverName,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// Test CALL TCP APIS
	testCallJWTURL(t, client, config)
}

func testCallJWTURL(t *testing.T, client *http.Client, config *Config) {
	addr := config.TCPAddress.String()
	url := fmt.Sprintf("https://%s/%s", addr, "jwt")
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	token := createToken(t, config.ServerCA)
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NotNil(t, body)

	assert.Equal(t, http.StatusOK, response.StatusCode)
	parsed, err := jwt.Parse(string(body), func(t *jwt.Token) (interface{}, error) { return config.ServerCA.PublicKey(), nil })
	require.NoError(t, err)
	require.NotNil(t, parsed)
}

func createToken(t *testing.T, CA ca.ServerCA) string {
	params := ca.JWTParams{
		Issuer:   "test-ca",
		Subject:  spiffeid.RequireTrustDomainFromString("domain.test"),
		Audience: []string{"galadriel-ca"},
		TTL:      time.Hour,
	}
	token, err := CA.SignJWT(params)
	require.NoError(t, err)

	return token
}

func newEndpointTestConfig(t *testing.T) (*Config, *x509.Certificate) {
	// used to generate a TCP address with a random port
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{})
	require.NoError(t, err)
	err = listener.Close()
	require.NoError(t, err)

	tempDir := t.TempDir()
	tcpAddr := listener.Addr().(*net.TCPAddr)
	localAddr := &net.UnixAddr{Net: "unix", Name: filepath.Join(tempDir, "sockets")}
	logger, _ := test.NewNullLogger()

	clk := clock.NewFake()
	clk.Set(time.Now())

	caCert, caKey, err := certtest.CreateTestCACertificate(clk)
	require.NoError(t, err)

	caConfig := &ca.Config{
		RootCert: caCert,
		RootKey:  caKey,
		Clock:    clk,
	}
	CA, err := ca.New(caConfig)
	require.NoError(t, err)

	config := &Config{
		TCPAddress:   tcpAddr,
		LocalAddress: localAddr,
		Logger:       logger,
		ServerCA:     CA,
		Clock:        clk,
		JWTTokenTTL:  time.Hour,
		X509CertTTL:  time.Hour,
	}

	return config, caCert
}

func waitForListening(t *testing.T, e *Endpoints, errCh chan error) {
	select {
	case <-e.hooks.tcpListening:
	case err := <-errCh:
		assert.Fail(t, err.Error())
	}
}
