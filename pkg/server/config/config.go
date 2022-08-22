package config

import (
	"fmt"
	"io"

	"github.com/hashicorp/hcl"
	"github.com/pkg/errors"
)

type ServerConfig struct {
	ServerConfigSection *ServerConfigSection `hcl:"server"`
}

type ServerConfigSection struct {
	SpireSocketPath string `hcl:"spire_socket_path"`
	ServerAddress   string `hcl:"server_address"`
	LogLevel        string `hcl:"log_level"`
}

func New(config io.Reader) (*ServerConfig, error) {

	if config == nil {
		return nil, errors.New("configuration is required")
	}

	configBytes, err := io.ReadAll(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read configuration")
	}

	return new(configBytes)
}

func new(configBytes []byte) (*ServerConfig, error) {
	var config ServerConfig

	if err := hcl.Decode(&config, string(configBytes)); err != nil {
		return nil, fmt.Errorf("unable to decode configuration: %v", err)
	}

	if config.ServerConfigSection == nil {
		config.ServerConfigSection = &ServerConfigSection{}
	}

	config.setDefaults()

	if err := config.validate(); err != nil {
		return nil, errors.Wrap(err, "invalid configuration")
	}

	return &config, nil
}

func (c *ServerConfig) validate() error {
	if c.ServerConfigSection.ServerAddress == "" {
		return errors.New("server.server_address is required")
	}

	return nil
}

func (c *ServerConfig) setDefaults() {
	if c.ServerConfigSection.LogLevel == "" {
		c.ServerConfigSection.LogLevel = "INFO"
	}

	if c.ServerConfigSection.SpireSocketPath == "" {
		c.ServerConfigSection.SpireSocketPath = "/tmp/spire-server/private/api.sock"
	}
}
