package cli

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/constants"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var hclConfigWithProviders = `server {
    listen_address = "127.0.0.1"
    listen_port = "2222"
    socket_path = "/tmp/api.sock"
	log_level = "DEBUG"
}

providers {
    Datastore "sqlite3" {
        connection_string = "./datastore.sqlite3"
    }

	X509CA "disk" {
		key_file_path = "./root_ca.key"
		cert_file_path = "./root_ca.crt"
	}

    KeyManager "memory" {}
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
			config: bytes.NewBuffer([]byte(hclConfigWithProviders)),
			expected: &Config{
				Server: &serverConfig{
					ListenAddress: "127.0.0.1",
					ListenPort:    2222,
					SocketPath:    "/tmp/api.sock",
					LogLevel:      "DEBUG",
				},
			},
		},
		{
			name:   "defaults",
			config: bytes.NewBuffer([]byte(`server {}`)),
			expected: &Config{
				Server: &serverConfig{
					ListenAddress: defaultAddress,
					ListenPort:    defaultPort,
					LogLevel:      constants.DefaultLogLevel,
				},
			},
		},
		{
			name:   "empty_config_file",
			config: bytes.NewBufferString(``),
			err:    "server section is empty",
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
			serverConfig, err := ParseConfig(tt.config)

			if tt.err != "" {
				assert.Nil(t, serverConfig)
				assert.Contains(t, err.Error(), tt.err)
				return
			}

			assert.Equal(t, tt.expected.Server, serverConfig.Server)
			assert.NoError(t, err)
		})
	}
}

func TestParseHCLConfigWithProviders(t *testing.T) {
	var config Config

	hclBody, err := hclsyntax.ParseConfig([]byte(hclConfigWithProviders), "", hcl.Pos{Line: 1, Column: 1})
	require.False(t, err.HasErrors())

	diagErr := gohcl.DecodeBody(hclBody.Body, nil, &config)
	require.False(t, diagErr.HasErrors())

	require.NotNil(t, config.Providers)
}
