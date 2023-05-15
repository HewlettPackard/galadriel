package cli

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var hclConfig = `harvester {
    local_socket_path = "/tmp/harvester/api.sock"
    spire_socket_path = "/tmp/api.sock"
    server_address = "localhost:7000"
    server_trust_bundle_path = "root_ca.crt"
    bundle_updates_interval = "1h"
    log_level = "DEBUG"
}
`

type fakeReader int

func (fakeReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("error from fake reader")
}

func TestNewHarvesterConfig(t *testing.T) {
	config := Config{Harvester: &harvesterConfig{
		LocalSocketPath:       "/tmp/harvester/api.sock",
		SpireSocketPath:       "/tmp/api.sock",
		ServerAddress:         "localhost:7000",
		ServerTrustBundlePath: "root_ca.crt",
		BundleUpdatesInterval: "10s",
		LogLevel:              "DEBUG",
	}}

	hc, err := NewHarvesterConfig(&config)
	require.NoError(t, err)

	assert.Equal(t, config.Harvester.LocalSocketPath, hc.LocalAddress.String())
	assert.Equal(t, "127.0.0.1", hc.ServerAddress.IP.String())
	assert.Equal(t, "127.0.0.1:7000", hc.ServerAddress.String())
	assert.Equal(t, config.Harvester.ServerTrustBundlePath, hc.ServerTrustBundlePath)
	assert.Equal(t, config.Harvester.BundleUpdatesInterval, hc.BundleUpdatesInterval.String())
	assert.Equal(t, config.Harvester.SpireSocketPath, hc.LocalSpireAddress.String())
	assert.Equal(t, "debug", strings.ToLower(config.Harvester.LogLevel))
}

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		config   io.Reader
		expected *Config
		err      string
	}{
		{
			name:   "ok",
			config: bytes.NewBuffer([]byte(hclConfig)),
			expected: &Config{
				Harvester: &harvesterConfig{
					LocalSocketPath:       "/tmp/harvester/api.sock",
					SpireSocketPath:       "/tmp/api.sock",
					ServerAddress:         "localhost:7000",
					ServerTrustBundlePath: "root_ca.crt",
					BundleUpdatesInterval: "1h",
					LogLevel:              "DEBUG",
				},
			},
		},
		{
			name: "defaults",
			config: bytes.NewBuffer([]byte(`harvester {
server_address = "localhost:5000"
server_trust_bundle_path = "./root_ca.crt"
}`)),
			expected: &Config{
				Harvester: &harvesterConfig{
					LocalSocketPath:       defaultLocalSocketPath,
					SpireSocketPath:       defaultSpireSocketPath,
					BundleUpdatesInterval: defaultBundleUpdatesInterval,
					LogLevel:              defaultLogLevel,
					ServerAddress:         "localhost:5000",
					ServerTrustBundlePath: "./root_ca.crt",
				},
			},
		},
		{
			name:   "empty_config_file",
			config: bytes.NewBufferString(``),
			err:    "harvester config section is empty",
		},
		{
			name:   "err_hcl",
			config: bytes.NewBufferString(`not a valid hcl`),
			err:    "failed to parse HCL",
		},
		{
			name:   "err_config_reader",
			config: nil,
			err:    "configuration is required",
		},
		{
			name:   "err_config_reader_error",
			config: fakeReader(0),
			err:    "failed to read configuration: error from fake reader",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			harvesterConfig, err := ParseConfig(tt.config)

			if tt.err != "" {
				assert.Nil(t, harvesterConfig)
				assert.Contains(t, err.Error(), tt.err)
				return
			}

			assert.Equal(t, tt.expected.Harvester, harvesterConfig.Harvester)
			assert.NoError(t, err)
		})
	}
}
