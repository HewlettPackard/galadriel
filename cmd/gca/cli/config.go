package cli

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/gca"
	"github.com/hashicorp/hcl"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	defaultSocketPath  = "/tmp/galadriel-ca/api.sock"
	defaultPort        = 8089
	defaultAddress     = "0.0.0.0"
	defaultLogLevel    = "INFO"
	defaultX509CertTTL = "1h"
	defaultJWTTokenTTL = "1h"
)

type Config struct {
	GCA *gcaConfig `hcl:"gca"`
}

type gcaConfig struct {
	ListenAddress string `hcl:"listen_address"`
	ListenPort    int    `hcl:"listen_port"`
	SocketPath    string `hcl:"socket_path"`
	LogLevel      string `hcl:"log_level"`
	RootCertPath  string `hcl:"root_cert_path"`
	RootKeyPath   string `hcl:"root_key_path"`
	X509CertTTL   string `hcl:"x509_cert_ttl"`
	JWTTokenTTL   string `hcl:"jwt_token_ttl"`
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

// NewGCAConfig creates a server.Config object from a cli.Config.
func NewGCAConfig(c *Config) (*gca.Config, error) {
	sc := &gca.Config{}

	addrPort := fmt.Sprintf("%s:%d", c.GCA.ListenAddress, c.GCA.ListenPort)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addrPort)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve TCP address: %w", err)
	}

	sc.TCPAddress = tcpAddr

	socketAddr, err := util.GetUnixAddrWithAbsPath(c.GCA.SocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to convert a socket_path value to a net.UnixAddr: %w", err)
	}

	sc.LocalAddress = socketAddr
	sc.Logger = logrus.WithField(telemetry.SubsystemName, telemetry.GaladrielCA)

	sc.RootCertPath = c.GCA.RootCertPath
	sc.RootKeyPath = c.GCA.RootKeyPath

	sc.X509CertTTL, err = time.ParseDuration(c.GCA.X509CertTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x509_cert_ttl config value: %s", c.GCA.X509CertTTL)
	}

	sc.JWTCertTTL, err = time.ParseDuration(c.GCA.JWTTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse jwt_token_ttl config value: %s", c.GCA.JWTTokenTTL)
	}

	return sc, nil
}

func newConfig(configBytes []byte) (*Config, error) {
	var config Config

	if err := hcl.Decode(&config, string(configBytes)); err != nil {
		return nil, fmt.Errorf("failed to decode configuration: %v", err)
	}

	if config.GCA == nil {
		return nil, errors.New("gca section is empty")
	}

	config.setDefaults()

	return &config, nil
}

func (c *Config) setDefaults() {
	if c.GCA.ListenAddress == "" {
		c.GCA.ListenAddress = defaultAddress
	}

	if c.GCA.ListenPort == 0 {
		c.GCA.ListenPort = defaultPort
	}

	if c.GCA.SocketPath == "" {
		c.GCA.SocketPath = defaultSocketPath
	}

	if c.GCA.LogLevel == "" {
		c.GCA.LogLevel = defaultLogLevel
	}

	if c.GCA.X509CertTTL == "" {
		c.GCA.X509CertTTL = defaultX509CertTTL
	}

	if c.GCA.JWTTokenTTL == "" {
		c.GCA.JWTTokenTTL = defaultJWTTokenTTL
	}
}
