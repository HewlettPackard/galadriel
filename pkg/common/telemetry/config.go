package telemetry

import (
	"fmt"
	"io"

	"github.com/hashicorp/hcl"
	"github.com/pkg/errors"
)

type TelemetryConfig struct {
	TelemetryConfigSection *TelemetryConfigSection `hcl:"telemetry"`
}

type TelemetryConfigSection struct {
	Prometheus *PrometheusConfig `hcl:"Prometheus"`
}

type PrometheusConfig struct {
	Host string `hcl:"host"`
	Port int    `hcl:"port"`
}

// New creates a new TelemetryConfig from the given input reader.
func New(config io.Reader) (*TelemetryConfig, error) {
	if config == nil {
		return nil, errors.New("configuration is required")
	}

	configBytes, err := io.ReadAll(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read configuration")
	}

	return new(configBytes)
}

func new(configBytes []byte) (*TelemetryConfig, error) {
	var config TelemetryConfig

	if err := hcl.Decode(&config, string(configBytes)); err != nil {
		return nil, fmt.Errorf("unable to decode configuration: %v", err)
	}

	if config.TelemetryConfigSection == nil {
		config.TelemetryConfigSection = &TelemetryConfigSection{}
	}

	return &config, nil
}
