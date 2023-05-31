package cli

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/constants"
	"github.com/stretchr/testify/assert"
)

var hclConfig = `harvester {
 	trust_domain = "example.org"	
    harvester_socket_path = "/tmp/harvester/api.sock"
    spire_socket_path = "/tmp/api.sock"
    galadriel_server_address = "localhost:7000"
    server_trust_bundle_path = "root_ca.crt"
    federated_bundles_poll_interval = "2h"
    spire_bundle_poll_interval = "1h"
    log_level = "DEBUG"
	data_dir = "/test"
}
`

type fakeReader int

func (fakeReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("error from fake reader")
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
					TrustDomain:                  "example.org",
					HarvesterSocketPath:          "/tmp/harvester/api.sock",
					SpireSocketPath:              "/tmp/api.sock",
					GaladrielServerAddress:       "localhost:7000",
					ServerTrustBundlePath:        "root_ca.crt",
					FederatedBundlesPollInterval: "2h",
					SpireBundlePollInterval:      "1h",
					LogLevel:                     "DEBUG",
					DataDir:                      "/test",
				},
			},
		},
		{
			name: "defaults",
			config: bytes.NewBuffer([]byte(`harvester {
trust_domain = "example.org"
harvester_socket_path = "/tmp/galadriel-harvester/api.sock"
spire_socket_path = "/tmp/spire-server/private/api.sock"
galadriel_server_address = "localhost:5000"
server_trust_bundle_path = "./root_ca.crt"
data_dir = "./data"
federated_bundles_poll_interval = "1h"
spire_bundle_poll_interval = "30m"
}`)),
			expected: &Config{
				Harvester: &harvesterConfig{
					TrustDomain:                  "example.org",
					HarvesterSocketPath:          "/tmp/galadriel-harvester/api.sock",
					SpireSocketPath:              "/tmp/spire-server/private/api.sock",
					GaladrielServerAddress:       "localhost:5000",
					ServerTrustBundlePath:        "./root_ca.crt",
					DataDir:                      "./data",
					FederatedBundlesPollInterval: "1h",
					SpireBundlePollInterval:      "30m",
					LogLevel:                     constants.DefaultLogLevel,
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
