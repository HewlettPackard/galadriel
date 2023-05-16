package cli

import (
	"fmt"
	"io"
	"net"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server"
	"github.com/HewlettPackard/galadriel/pkg/server/catalog"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	defaultSocketPath = "/tmp/galadriel-server/api.sock"
	defaultPort       = 8085
	defaultAddress    = "0.0.0.0"
	defaultLogLevel   = "INFO"
)

// Config holds the configuration for the Galadriel server.
type Config struct {
	Server    *serverConfig   `hcl:"server,block"`
	Providers *providersBlock `hcl:"providers,block"`
}

type serverConfig struct {
	ListenAddress string `hcl:"listen_address,optional"`
	ListenPort    int    `hcl:"listen_port,optional"`
	SocketPath    string `hcl:"socket_path,optional"`
	LogLevel      string `hcl:"log_level,optional"`
	DBConnString  string `hcl:"db_conn_string"`
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

	sc.DBConnString = c.Server.DBConnString

	logLevel, err := logrus.ParseLevel(c.Server.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %v", err)
	}
	logger := logrus.New()
	logger.SetLevel(logLevel)
	sc.Logger = logger.WithField(telemetry.SubsystemName, telemetry.Server)

	// TODO: eventually providers section will be required
	if c.Providers != nil {
		sc.ProvidersConfig, err = catalog.ProvidersConfigsFromHCLBody(c.Providers.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to parse providers configuration: %w", err)
		}
	}

	return sc, nil
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

	if config.Server == nil {
		return nil, errors.New("server section is empty")
	}

	config.setDefaults()

	return &config, nil
}

func (c *Config) setDefaults() {
	if c.Server.ListenAddress == "" {
		c.Server.ListenAddress = defaultAddress
	}

	if c.Server.ListenPort == 0 {
		c.Server.ListenPort = defaultPort
	}

	if c.Server.SocketPath == "" {
		c.Server.SocketPath = defaultSocketPath
	}

	if c.Server.LogLevel == "" {
		c.Server.LogLevel = defaultLogLevel
	}
}
