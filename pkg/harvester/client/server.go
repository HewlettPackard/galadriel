package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/constants"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/sirupsen/logrus"
)

const (
	jsonContentType = "application/json"

	postBundlePath     = "/bundle"
	postBundleSyncPath = "/bundle/sync"
	onboardPath        = "/trust-domain/onboard"
	jwtPath            = "/trust-domain/jwt"
)

// GaladrielServerClient represents a client to connect to Galadriel Server
type GaladrielServerClient interface {
	SyncFederatedBundles(context.Context, *common.SyncBundleRequest) (*common.SyncBundleResponse, error)
	PostBundle(context.Context, *common.PostBundleRequest) error
	Onboard(ctx context.Context, token string) error
	GetNewJWTToken(ctx context.Context) error
}

type client struct {
	httpClient  *http.Client
	address     *net.TCPAddr
	logger      logrus.FieldLogger
	jwtProvider *jwtProvider
	errChan     chan error
}

type jwtProvider struct {
	mu  sync.RWMutex
	jwt string
}

// NewGaladrielServerClient creates a new Galadriel Server client, using the given token to authenticate
// and the given trustBundlePath to validate the server certificate.
func NewGaladrielServerClient(address *net.TCPAddr, trustBundlePath string) (GaladrielServerClient, error) {
	skipOnboard := func(req *http.Request) bool {
		return strings.Contains(req.URL.Path, onboardPath)
	}

	jp := &jwtProvider{}

	c, err := createTLSClient(trustBundlePath, jp, skipOnboard)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS client: %w", err)
	}

	errChan := make(chan error)
	client := &client{
		httpClient:  c,
		address:     address,
		logger:      logrus.WithField(telemetry.SubsystemName, telemetry.GaladrielServerClient),
		jwtProvider: jp,
		errChan:     errChan,
	}

	return client, nil
}

func (c *client) Onboard(ctx context.Context, token string) error {
	url := fmt.Sprintf("%s%s?joinToken=%s", c.getHTTPAddress(), onboardPath, token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyString, err := readBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}
		return errors.New(bodyString)
	}

	bodyString, err := readBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if bodyString == "" {
		return errors.New("empty response body")
	}
	c.jwtProvider.setToken(bodyString)

	c.logger.Info("Connected to Galadriel Server")

	return nil
}

func (c *client) GetNewJWTToken(ctx context.Context) error {
	url := fmt.Sprintf("%s%s", c.getHTTPAddress(), jwtPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyString, err := readBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}
		return errors.New(bodyString)
	}

	bodyString, err := readBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if bodyString == "" {
		return errors.New("empty response body")
	}

	c.jwtProvider.setToken(bodyString)

	return nil
}

func (c *client) SyncFederatedBundles(ctx context.Context, req *common.SyncBundleRequest) (*common.SyncBundleResponse, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal federated bundle request: %v", err)
	}

	c.logger.Debugf("Sending post federated bundles updates:\n%s", b)
	url := c.address.String() + postBundleSyncPath
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	res, err := c.httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned an error code %d: \n%s", res.StatusCode, body)
	}

	var syncBundleResponse common.SyncBundleResponse
	if err := json.Unmarshal(body, &syncBundleResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sync bundle response: %v", err)
	}

	return &syncBundleResponse, nil
}

func (c *client) PostBundle(ctx context.Context, req *common.PostBundleRequest) error {
	b, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal push bundle request: %v", err)
	}

	url := c.address.String() + postBundlePath

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create push bundle request: %v", err)
	}

	res, err := c.httpClient.Do(r)
	if err != nil {
		return fmt.Errorf("failed to send push bundle request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("push bundle request returned an error code %d: \n%s", res.StatusCode, body)
	}

	return nil
}

func createTLSClient(trustBundlePath string, jwtProvider *jwtProvider, skipper func(*http.Request) bool) (*http.Client, error) {
	caCert, err := os.ReadFile(trustBundlePath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		return nil, fmt.Errorf("failed to append CA certificates")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    caCertPool,
			ServerName: constants.GaladrielServerName,
		},
	}

	return &http.Client{
		Transport: &decoratedTransport{
			jwtProvider: jwtProvider,
			transport:   transport,
			skipper:     skipper,
		},
	}, nil
}

type decoratedTransport struct {
	jwtProvider *jwtProvider
	transport   *http.Transport
	skipper     func(*http.Request) bool
}

// RoundTrip applies the decorator to every request adding the Authorization header
func (t *decoratedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.skipper != nil && t.skipper(req) {
		return t.transport.RoundTrip(req)
	}

	// Apply decorator
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.jwtProvider.getToken()))
	req.Header.Set("Content-Type", jsonContentType)

	return t.transport.RoundTrip(req)
}

func (j *jwtProvider) setToken(t string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	// Sanitize token removing leading and trailing spaces and quotes
	token := strings.TrimSpace(t)
	token = strings.ReplaceAll(token, "\"", "")
	j.jwt = token
}

func (j *jwtProvider) getToken() string {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.jwt
}

func (c *client) getHTTPAddress() string {
	return fmt.Sprintf("https://%s", c.address.String())
}

func readBody(resp *http.Response) (string, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	return bodyString, nil
}
