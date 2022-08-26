package config

import (
	"fmt"
	"io"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/hashicorp/hcl"
	"github.com/pkg/errors"
)

type HarvesterConfig struct {
	HarvesterConfigSection *HarvesterConfigSection           `hcl:"harvester"`
	TelemetryConfigSection *telemetry.TelemetryConfigSection `hcl:"telemetry"`
}

type HarvesterConfigSection struct {
	SpireSocketPath string `hcl:"spire_socket_path"`
	ServerAddress   string `hcl:"server_address"`
	LogLevel        string `hcl:"log_level"`
}

// New creates a new HarvesterConfig from the given input reader.
func New(config io.Reader) (*HarvesterConfig, error) {

	if config == nil {
		return nil, errors.New("configuration is required")
	}

	configBytes, err := io.ReadAll(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read configuration")
	}

	return new(configBytes)
}

func new(configBytes []byte) (*HarvesterConfig, error) {
	var config HarvesterConfig

	if err := hcl.Decode(&config, string(configBytes)); err != nil {
		return nil, fmt.Errorf("unable to decode configuration: %v", err)
	}

	if config.HarvesterConfigSection == nil {
		config.HarvesterConfigSection = &HarvesterConfigSection{}
		config.TelemetryConfigSection = &telemetry.TelemetryConfigSection{}
	}

	config.setDefaults()

	if err := config.validate(); err != nil {
		return nil, errors.Wrap(err, "bad configuration")
	}

	return &config, nil
}

func (c *HarvesterConfig) validate() error {
	if c.HarvesterConfigSection.ServerAddress == "" {
		return errors.New("harvester.server_address is required")
	}

	return nil
}

func (c *HarvesterConfig) setDefaults() {
	if c.HarvesterConfigSection.LogLevel == "" {
		c.HarvesterConfigSection.LogLevel = "INFO"
	}

	if c.HarvesterConfigSection.SpireSocketPath == "" {
		c.HarvesterConfigSection.SpireSocketPath = "/tmp/spire-server/private/api.sock"
	}
}
