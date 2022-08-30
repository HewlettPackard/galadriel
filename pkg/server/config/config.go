package config

import (
	"fmt"
	"io"

	"github.com/hashicorp/hcl"
	"github.com/pkg/errors"
)

type Server struct {
	ServerConfigSection *ServerConfigSection `hcl:"server"`
}

type ServerConfigSection struct {
	ListenAddress string `hcl:"listen_address"`
	LogLevel      string `hcl:"log_level"`
}

func New(config io.Reader) (*Server, error) {

	if config == nil {
		return nil, errors.New("configuration is required")
	}

	configBytes, err := io.ReadAll(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read configuration")
	}

	return newConfig(configBytes)
}

func newConfig(configBytes []byte) (*Server, error) {
	var config Server

	if err := hcl.Decode(&config, string(configBytes)); err != nil {
		return nil, fmt.Errorf("unable to decode configuration: %v", err)
	}

	if config.ServerConfigSection == nil {
		return nil, errors.Wrap(errors.New("configuration file is empty"), "bad configuration")
	}

	config.setDefaults()

	return &config, nil
}

func (c *Server) setDefaults() {
	if c.ServerConfigSection.ListenAddress == "" {
		c.ServerConfigSection.ListenAddress = "localhost:8080"
	}

	if c.ServerConfigSection.LogLevel == "" {
		c.ServerConfigSection.LogLevel = "INFO"
	}
}
