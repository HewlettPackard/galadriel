package cli

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester"
	"github.com/hashicorp/hcl"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	defaultSpireSocketPath       = "/tmp/spire-server/private/api.sock"
	defaultBundleUpdatesInterval = "30s"
	defaultLogLevel              = "INFO"
)

// TODO: migrate to HCL 2 (see Server CLI)

type Config struct {
	Harvester *harvesterConfig `hcl:"harvester"`
}

type harvesterConfig struct {
	SpireSocketPath       string `hcl:"spire_socket_path"`
	ServerAddress         string `hcl:"server_address"`
	ServerTrustBundlePath string `hcl:"server_trust_bundle_path"`
	BundleUpdatesInterval string `hcl:"bundle_updates_interval"`
	LogLevel              string `hcl:"log_level"`
}

// ParseConfig reads a configuration from the Reader and parses it
// to a cli.Config object setting the defaults for the missing values.
func ParseConfig(config io.Reader) (*Config, error) {
	if config == nil {
		return nil, errors.New("configuration is required")
	}

	configBytes, err := io.ReadAll(config)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration: %w", err)
	}

	return newConfig(configBytes)
}

// NewHarvesterConfig creates a harvester.Config object from a cli.Config.
func NewHarvesterConfig(c *Config) (*harvester.Config, error) {
	hc := &harvester.Config{}

	spireAddr, err := util.GetUnixAddrWithAbsPath(c.Harvester.SpireSocketPath)
	if err != nil {
		return nil, err
	}

	buInt, err := time.ParseDuration(c.Harvester.BundleUpdatesInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bundle updates interval: %v", err)
	}

	serverTCPAddress, err := net.ResolveTCPAddr("tcp", c.Harvester.ServerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve server address: %v", err)
	}

	hc.SpireAddress = spireAddr
	hc.ServerAddress = serverTCPAddress
	hc.ServerTrustBundlePath = c.Harvester.ServerTrustBundlePath
	hc.BundleUpdatesInterval = buInt

	hc.Logger = logrus.WithField(telemetry.SubsystemName, telemetry.Harvester)

	return hc, nil
}

func newConfig(configBytes []byte) (*Config, error) {
	var config Config

	if err := hcl.Decode(&config, string(configBytes)); err != nil {
		return nil, fmt.Errorf("unable to decode configuration: %w", err)
	}

	if config.Harvester == nil {
		return nil, errors.New("harvester section is empty")
	}

	config.setDefaults()

	return &config, nil
}

func (c *Config) setDefaults() {
	if c.Harvester.SpireSocketPath == "" {
		c.Harvester.SpireSocketPath = defaultSpireSocketPath
	}

	if c.Harvester.BundleUpdatesInterval == "" {
		c.Harvester.BundleUpdatesInterval = defaultBundleUpdatesInterval
	}

	if c.Harvester.LogLevel == "" {
		c.Harvester.LogLevel = defaultLogLevel
	}
}
