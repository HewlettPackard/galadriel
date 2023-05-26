package cli

import (
	"fmt"
	"io"
	"net"

	"github.com/HewlettPackard/galadriel/pkg/common/constants"
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
	// TODO: These defaults should be moved close to where they are used (Server, Endpoints).
	defaultPort    = 8085
	defaultAddress = "0.0.0.0"
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
// It returns a *server.Config and an error if any.
func NewServerConfig(c *Config) (*server.Config, error) {
	sc := &server.Config{}

	addrPort := fmt.Sprintf("%s:%d", c.Server.ListenAddress, c.Server.ListenPort)
	tcpAddr, err := net.ResolveTCPAddr(constants.TCPProtocol, addrPort)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve TCP address %s: %w", addrPort, err)
	}

	sc.TCPAddress = tcpAddr

	socketAddr, err := util.GetUnixAddrWithAbsPath(c.Server.SocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get Unix address from path %s: %w", c.Server.SocketPath, err)
	}

	sc.LocalAddress = socketAddr

	logLevel, err := logrus.ParseLevel(c.Server.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level %s: %w", c.Server.LogLevel, err)
	}
	logger := logrus.New()
	logger.SetLevel(logLevel)
	sc.Logger = logger.WithField(telemetry.SubsystemName, telemetry.Server)

	sc.ProvidersConfig, err = catalog.ProvidersConfigsFromHCLBody(c.Providers.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse providers configuration: %w", err)
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

	if c.Server.LogLevel == "" {
		c.Server.LogLevel = constants.DefaultLogLevel
	}
}
