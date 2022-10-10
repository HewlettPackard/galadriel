package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/sirupsen/logrus"
)

const (
	contentType = "application/json"

	postBundlePath     = "/bundle"
	postBundleSyncPath = "/bundle/sync"
	onboardPath        = "/onboard"
)

// GaladrielServerClient represents a client to connect to Galadriel Server
type GaladrielServerClient interface {
	SyncFederatedBundles(context.Context, *common.SyncBundleRequest) (*common.SyncBundleResponse, error)
	PostBundle(context.Context, *common.PostBundleRequest) error
	Connect(ctx context.Context, token string) error
}

type client struct {
	c       http.Client
	address string
	token   string
	logger  logrus.FieldLogger
}

func NewGaladrielServerClient(address, token string) (GaladrielServerClient, error) {
	return &client{
		c:       *http.DefaultClient,
		address: "http://" + address,
		token:   token,
		logger:  logrus.WithField(telemetry.SubsystemName, telemetry.GaladrielServerClient),
	}, nil
}

func (c *client) Connect(ctx context.Context, token string) error {
	url := c.address + onboardPath
	req, err := http.NewRequestWithContext(ctx, http.MethodConnect, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyString, err := readBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}
		return fmt.Errorf("failed to connect to Galadriel Server: %s", bodyString)
	}

	c.logger.Info("Connected to Galadriel Server")
	return nil
}

func (c *client) SyncFederatedBundles(ctx context.Context, req *common.SyncBundleRequest) (*common.SyncBundleResponse, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal federated bundle request: %v", err)
	}

	c.logger.Debugf("Sending post federated bundles updates:\n%s", b)
	url := c.address + postBundleSyncPath
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// TODO: decorate all requests coming out
	r.Header.Set("Authorization", "Bearer "+c.token)
	r.Header.Set("Content-Type", contentType)

	res, err := c.c.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// TODO: check right status code
	if res.StatusCode != 200 {
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

	url := c.address + postBundlePath
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create push bundle request: %v", err)
	}

	// TODO: decorate all requests coming out
	r.Header.Set("Authorization", "Bearer "+c.token)
	r.Header.Set("Content-Type", contentType)

	res, err := c.c.Do(r)
	if err != nil {
		return fmt.Errorf("failed to send push bundle request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// TODO: check right status code
	if res.StatusCode != 200 {
		return fmt.Errorf("push bundle request returned an error code %d: \n%s", res.StatusCode, body)
	}

	return nil
}

func readBody(resp *http.Response) (string, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	return bodyString, nil
}
