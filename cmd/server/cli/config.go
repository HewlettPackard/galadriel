package cli

import (
	"fmt"
	"io"
	"net"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server"
	"github.com/hashicorp/hcl"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	defaultSocketPath = "/tmp/galadriel-server/api.sock"
)

type Config struct {
	Server *serverConfig `hcl:"server"`
}

type serverConfig struct {
	ListenAddress string `hcl:"listen_address"`
	ListenPort    int    `hcl:"listen_port"`
	SocketPath    string `hcl:"socket_path"`
	LogLevel      string `hcl:"log_level"`
	DBConnString  string `hcl:"db_conn_string"`
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

// NewServerConfig creates a server.Config object from a cli.Config.
func NewServerConfig(c *Config) (*server.Config, error) {
	sc := &server.Config{}

	addrPort := fmt.Sprintf("%s:%d", c.Server.ListenAddress, c.Server.ListenPort)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addrPort)
	if err != nil {
		return nil, err
	}

	sc.TCPAddress = tcpAddr

	socketAddr, err := util.GetUnixAddrWithAbsPath(c.Server.SocketPath)
	if err != nil {
		return nil, err
	}

	sc.LocalAddress = socketAddr
	sc.Logger = logrus.WithField(telemetry.SubsystemName, telemetry.GaladrielServer)

	sc.DBConnString = c.Server.DBConnString

	return sc, nil
}

func newConfig(configBytes []byte) (*Config, error) {
	var config Config

	if err := hcl.Decode(&config, string(configBytes)); err != nil {
		return nil, fmt.Errorf("unable to decode configuration: %v", err)
	}

	if config.Server == nil {
		return nil, errors.New("server section is empty")
	}

	config.setDefaults()

	return &config, nil
}

func (c *Config) setDefaults() {
	if c.Server.ListenAddress == "" {
		c.Server.ListenAddress = "0.0.0.0"
	}

	if c.Server.ListenPort == 0 {
		c.Server.ListenPort = 8085
	}

	if c.Server.SocketPath == "" {
		c.Server.SocketPath = defaultSocketPath
	}

	if c.Server.LogLevel == "" {
		c.Server.LogLevel = "INFO"
	}
}
