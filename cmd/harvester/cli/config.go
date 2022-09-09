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
	"net"
)

const (
	defaultSocketPath      = "/tmp/galadriel-harvester/api.sock"
	defaultSpireSocketPath = "/tmp/spire-server/private/api.sock"
)

type Config struct {
	Harvester *harvesterConfig `hcl:"harvester"`
}

type harvesterConfig struct {
	ListenAddress   string `hcl:"listen_address"`
	ListenPort      int    `hcl:"listen_port"`
	SocketPath      string `hcl:"socket_path"`
	SpireSocketPath string `hcl:"spire_socket_path"`
	ServerAddress   string `hcl:"server-address"`
	LogLevel        string `hcl:"log_level"`
}

// ParseConfig reads a configuration from the Reader and parses it
// to a cli.Config object setting the defaults for the missing values.
func ParseConfig(config io.Reader) (*Config, error) {

	if config == nil {
		return nil, errors.New("configuration is required")
	}

	configBytes, err := io.ReadAll(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read configuration")
	}

	return newConfig(configBytes)
}

// NewHarvesterConfig creates a harvester.Config object from a cli.Config.
func NewHarvesterConfig(c *Config) (*harvester.Config, error) {
	sc := &harvester.Config{}

	ip := net.ParseIP(c.Harvester.ListenAddress)
	bindAddr := &net.TCPAddr{
		IP:   ip,
		Port: c.Harvester.ListenPort,
	}
	sc.TCPAddress = bindAddr

	socketAddr, err := util.GetUnixAddrWithAbsPath(c.Harvester.SocketPath)
	if err != nil {
		return nil, err
	}

	sc.LocalAddress = socketAddr

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
		return nil, fmt.Errorf("unable to decode configuration: %v", err)
	}

	if config.Harvester == nil {
		return nil, errors.Wrap(errors.New("configuration file is empty"), "bad configuration")
	}

	config.setDefaults()

	return &config, nil
}

func (c *Config) setDefaults() {
	if c.Harvester.ListenAddress == "" {
		c.Harvester.ListenAddress = "0.0.0.0"
	}

	if c.Harvester.ListenPort == 0 {
		c.Harvester.ListenPort = 8086
	}

	if c.Harvester.SocketPath == "" {
		c.Harvester.SocketPath = defaultSocketPath
	}

	if c.Harvester.SpireSocketPath == "" {
		c.Harvester.SpireSocketPath = defaultSpireSocketPath
	}

	if c.Harvester.LogLevel == "" {
		c.Harvester.LogLevel = "INFO"
	}
}
