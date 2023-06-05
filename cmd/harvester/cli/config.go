package cli

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/constants"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester"
	"github.com/HewlettPackard/galadriel/pkg/harvester/catalog"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type Config struct {
	Harvester *harvesterConfig `hcl:"harvester,block"`
	Providers *providersBlock  `hcl:"providers,block"`
}

type harvesterConfig struct {
	TrustDomain                  string `hcl:"trust_domain"`
	HarvesterSocketPath          string `hcl:"harvester_socket_path,optional"`
	SpireSocketPath              string `hcl:"spire_socket_path,optional"`
	GaladrielServerAddress       string `hcl:"galadriel_server_address"`
	ServerTrustBundlePath        string `hcl:"server_trust_bundle_path"`
	FederatedBundlesPollInterval string `hcl:"federated_bundles_poll_interval,optional"`
	SpireBundlePollInterval      string `hcl:"spire_bundle_poll_interval,optional"`
	LogLevel                     string `hcl:"log_level,optional"`
	DataDir                      string `hcl:"data_dir"`
}

// providersBlock holds the Providers HCL block body.
type providersBlock struct {
	Body hcl.Body `hcl:",remain"`
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

	trustDomain, err := spiffeid.TrustDomainFromString(c.Harvester.TrustDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trust domain: %v", err)
	}
	hc.TrustDomain = trustDomain

	localAddr, err := util.GetUnixAddrWithAbsPath(c.Harvester.HarvesterSocketPath)
	if err != nil {
		return nil, err
	}
	hc.HarvesterSocketPath = localAddr

	spireAddr, err := util.GetUnixAddrWithAbsPath(c.Harvester.SpireSocketPath)
	if err != nil {
		return nil, err
	}
	hc.SpireSocketPath = spireAddr

	if c.Harvester.FederatedBundlesPollInterval != "" {
		federatedBundlesPollInterval, err := time.ParseDuration(c.Harvester.FederatedBundlesPollInterval)
		if err != nil {
			return nil, fmt.Errorf("failed to parse federated bundles poll interval: %v", err)
		}
		hc.FederatedBundlesPollInterval = federatedBundlesPollInterval
	}

	if c.Harvester.SpireBundlePollInterval != "" {
		spireBundlePollInterval, err := time.ParseDuration(c.Harvester.SpireBundlePollInterval)
		if err != nil {
			return nil, fmt.Errorf("failed to parse spire bundle poll interval: %v", err)
		}
		hc.SpireBundlePollInterval = spireBundlePollInterval
	}

	serverTCPAddress, err := net.ResolveTCPAddr(constants.TCPProtocol, c.Harvester.GaladrielServerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve server address: %v", err)
	}
	hc.GaladrielServerAddress = serverTCPAddress

	logLevel, err := logrus.ParseLevel(c.Harvester.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %v", err)
	}
	logger := logrus.New()
	logger.SetLevel(logLevel)
	hc.Logger = logger.WithField(telemetry.SubsystemName, telemetry.Harvester)

	hc.DataDir = c.Harvester.DataDir
	hc.ServerTrustBundlePath = c.Harvester.ServerTrustBundlePath

	hc.ProvidersConfig, err = catalog.ProvidersConfigsFromHCLBody(c.Providers.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse providers configuration: %v", err)
	}

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
	if config.HarvesterSocketPath == "" {
		config.HarvesterSocketPath = defaultSocketPath
	}

	if config.LogLevel == "" {
		config.LogLevel = constants.DefaultLogLevel
	}
}
