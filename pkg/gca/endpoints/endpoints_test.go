package endpoints

import (
	"context"
	"crypto/x509"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/ca"
	"github.com/HewlettPackard/galadriel/test/certtest"
	"github.com/jmhodges/clock"
	"github.com/sirupsen/logrus/hooks/test"
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
	config, _ := newEndpointTestConfig(t)

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
