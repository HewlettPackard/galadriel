package endpoints

import (
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/keymanager"
	"github.com/HewlettPackard/galadriel/pkg/common/x509ca"
	"github.com/HewlettPackard/galadriel/pkg/common/x509ca/disk"
	"github.com/HewlettPackard/galadriel/pkg/server/db"
	"github.com/HewlettPackard/galadriel/test/certtest"
	"github.com/jmhodges/clock"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCatalog struct {
	ds         db.Datastore
	x509ca     x509ca.X509CA
	keyManager keymanager.KeyManager
}

func (c fakeCatalog) GetX509CA() x509ca.X509CA {
	return c.x509ca
}

func (c fakeCatalog) GetKeyManager() keymanager.KeyManager {
	return c.keyManager
}

func (c fakeCatalog) GetDatastore() db.Datastore {
	return c.ds
}

func TestListenAndServe(t *testing.T) {
	config := newEndpointTestConfig(t)

	endpoints, err := New(config)
	require.NoError(t, err)

	endpoints.hooks.tcpListening = make(chan struct{})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

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

func newEndpointTestConfig(t *testing.T) *Config {
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

	certsFolder := certtest.CreateTestCACertificates(t, clk)

	ca, err := disk.New()
	require.NoError(t, err)
	c := &disk.Config{
		CertFilePath: certsFolder + "/root-ca.crt",
		KeyFilePath:  certsFolder + "/root-ca.key",
	}
	err = ca.Configure(c)
	require.NoError(t, err)

	km := keymanager.NewMemoryKeyManager(nil)

	cat := fakeCatalog{
		x509ca:     ca,
		keyManager: km,
	}

	config := &Config{
		TCPAddress:   tcpAddr,
		LocalAddress: localAddr,
		Logger:       logger,
		Catalog:      cat,
	}

	return config
}

func waitForListening(t *testing.T, e *Endpoints, errCh chan error) {
	select {
	case <-e.hooks.tcpListening:
	case err := <-errCh:
		t.Fatalf("Failed to start Endpoints: %v", err)
	}
}
