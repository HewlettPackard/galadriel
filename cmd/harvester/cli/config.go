package cli

import (
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/harvester"
	"github.com/hashicorp/hcl"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
)

const (
	defaultSpireSocketPath = "/tmp/spire-server/private/api.sock"
)

type Config struct {
	Harvester *harvesterConfig `hcl:"harvester"`
}

type harvesterConfig struct {
	SpireSocketPath string `hcl:"spire_socket_path"`
	ServerAddress   string `hcl:"server_address"`
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
	sc := &harvester.Config{}

	spireAddr, err := util.GetUnixAddrWithAbsPath(c.Harvester.SpireSocketPath)
	if err != nil {
		return nil, err
	}

	sc.SpireAddress = spireAddr

	sc.ServerAddress = c.Harvester.ServerAddress

	sc.Log = logrus.WithField(telemetry.SubsystemName, telemetry.Harvester)

	return sc, nil
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
}
