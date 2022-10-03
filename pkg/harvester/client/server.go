package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/sirupsen/logrus"
)

const (
	contentType = "application/json"

	postBundlePath = "/bundle"
)

// GaladrielServerClient represents a client to connect to Galadriel Server
type GaladrielServerClient interface {
	GetUpdates(context.Context) ([]string, error)
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
		address: address,
		token:   token,
		logger:  logrus.WithField(telemetry.SubsystemName, telemetry.GaladrielServerClient),
	}, nil
}

func (c *client) Connect(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, "CONNECT", fmt.Sprintf("http://%s/onboard", c.address), nil)
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

func readBody(resp *http.Response) (string, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	return bodyString, nil
}

func (c *client) GetUpdates(ctx context.Context) ([]string, error) {
	return nil, errors.New("not implemented")
}

func (s *client) PostBundle(ctx context.Context, req *common.PostBundleRequest) error {
	b, err := json.Marshal(req)
	if err != nil {
		s.logger.Debug("Pushing Bundle: \n" + string(b))
		return fmt.Errorf("failed to marshal push bundle request: %v", err)
	}

	url := s.address + postBundlePath
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create push bundle request: %v", err)
	}

	// TODO: decorate all requests coming out
	r.Header.Set("Authorization", "Bearer "+s.token)
	r.Header.Set("Content-Type", contentType)

	res, err := s.c.Do(r)
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
