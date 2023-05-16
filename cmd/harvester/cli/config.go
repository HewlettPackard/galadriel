package cli

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	defaultLocalSocketPath       = "/tmp/galadriel-harvester/api.sock"
	defaultSpireSocketPath       = "/tmp/spire-server/private/api.sock"
	defaultBundleUpdatesInterval = "30s"
	defaultLogLevel              = "INFO"
)

type Config struct {
	Harvester *harvesterConfig `hcl:"harvester,block"`
}

type harvesterConfig struct {
	LocalSocketPath       string `hcl:"local_socket_path,optional"`
	SpireSocketPath       string `hcl:"spire_socket_path,optional"`
	ServerAddress         string `hcl:"server_address"`
	ServerTrustBundlePath string `hcl:"server_trust_bundle_path"`
	BundleUpdatesInterval string `hcl:"bundle_updates_interval,optional"`
	LogLevel              string `hcl:"log_level,optional"`
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

	if c.Harvester == nil {
		return nil, errors.New("harvester configuration is required")
	}

	localAddr, err := util.GetUnixAddrWithAbsPath(c.Harvester.LocalSocketPath)
	if err != nil {
		return nil, err
	}

	spireAddr, err := util.GetUnixAddrWithAbsPath(c.Harvester.SpireSocketPath)
	if err != nil {
		return nil, err
	}

	bundleInterval, err := time.ParseDuration(c.Harvester.BundleUpdatesInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bundle updates interval: %v", err)
	}

	serverTCPAddress, err := net.ResolveTCPAddr("tcp", c.Harvester.ServerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve server address: %v", err)
	}

	hc.LocalAddress = localAddr
	hc.LocalSpireAddress = spireAddr
	hc.ServerAddress = serverTCPAddress
	hc.ServerTrustBundlePath = c.Harvester.ServerTrustBundlePath
	hc.BundleUpdatesInterval = bundleInterval

	logLevel, err := logrus.ParseLevel(c.Harvester.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %v", err)
	}
	logger := logrus.New()
	logger.SetLevel(logLevel)
	hc.Logger = logger.WithField(telemetry.SubsystemName, telemetry.Harvester)

	return hc, nil
}

func newConfig(configBytes []byte) (*Config, error) {
	var config Config

	hclBody, err := hclsyntax.ParseConfig(configBytes, "", hcl.Pos{Line: 1, Column: 1})
	if err != nil {
		return nil, fmt.Errorf("failed to parse HCL: %w", err)
	}

	if err := gohcl.DecodeBody(hclBody.Body, nil, &config); err != nil {
		return nil, fmt.Errorf("failed to decode HCL: %w", err)
	}

	if config.Harvester == nil {
		return nil, errors.New("harvester config section is empty")
	}

	setDefaults(config.Harvester)

	return &config, nil
}

func setDefaults(config *harvesterConfig) {
	if config.LocalSocketPath == "" {
		config.LocalSocketPath = defaultLocalSocketPath
	}

	if config.SpireSocketPath == "" {
		config.SpireSocketPath = defaultSpireSocketPath
	}

	if config.BundleUpdatesInterval == "" {
		config.BundleUpdatesInterval = defaultBundleUpdatesInterval
	}

	if config.LogLevel == "" {
		config.LogLevel = defaultLogLevel
	}
}
