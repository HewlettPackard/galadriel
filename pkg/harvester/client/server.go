package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// GaladrielServerClient represents a client to connect to Galadriel Server
type GaladrielServerClient interface {
	GetUpdates(context.Context) ([]string, error)
	PushUpdates(context.Context, []string) error
	Connect(ctx context.Context, token string) error
}

type client struct {
	address string
	c       http.Client
	logger  logrus.FieldLogger
}

func NewGaladrielServerClient(address string) (GaladrielServerClient, error) {
	c := http.Client{
		Transport: http.DefaultTransport,
	}

	return &client{
		c:       c,
		address: address,
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

func (c *client) PushUpdates(ctx context.Context, updates []string) error {
	return errors.New("not implemented")
}
