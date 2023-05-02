package cli

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var hclConfigWithProviders = `server {
    listen_address = "127.0.0.1"
    listen_port = "2222"
    socket_path = "/tmp/api.sock"
    db_conn_string = "test_conn_string" 
	log_level = "DEBUG"
}

providers {
	x509ca "disk" {
		key_file_path = "./root_ca.key"
		cert_file_path = "./root_ca.crt"
	}
}
`

type fakeReader int

func (fakeReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("error from fake reader")
}

func TestNewServerConfig(t *testing.T) {
	config := Config{Server: &serverConfig{
		ListenAddress: "localhost",
		ListenPort:    8000,
		SocketPath:    "/example",
		LogLevel:      "INFO",
		DBConnString:  "postgresql://postgres:postgres@localhost:5432/galadriel",
	}}

	sc, err := NewServerConfig(&config)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "127.0.0.1", sc.TCPAddress.IP.String())
	assert.Equal(t, config.Server.ListenPort, sc.TCPAddress.Port)
	assert.Equal(t, config.Server.SocketPath, sc.LocalAddress.String())
	assert.Equal(t, strings.ToLower(config.Server.LogLevel), logrus.GetLevel().String())
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
					DBConnString:  "test_conn_string",
				},
			},
		},
		{
			name: "defaults",
			config: bytes.NewBuffer([]byte(`server {
db_conn_string = "test_conn_string"
}`)),
			expected: &Config{
				Server: &serverConfig{
					ListenAddress: defaultAddress,
					ListenPort:    defaultPort,
					SocketPath:    defaultSocketPath,
					LogLevel:      defaultLogLevel,
					DBConnString:  "test_conn_string",
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
