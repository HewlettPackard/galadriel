package cli

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

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
			name: "ok",
			config: bytes.NewBuffer([]byte(
				`server { listen_address = "127.0.0.1" listen_port = 2222 socket_path = "/tmp/api.sock" log_level = "INFO" db_conn_string = "postgresql://postgres:postgres@localhost:5432/galadriel"}`)),
			expected: &Config{
				Server: &serverConfig{
					ListenAddress: "127.0.0.1",
					ListenPort:    2222,
					SocketPath:    "/tmp/api.sock",
					LogLevel:      "INFO",
					DBConnString:  "postgresql://postgres:postgres@localhost:5432/galadriel",
				},
			},
		},
		{
			name:   "defaults",
			config: bytes.NewBuffer([]byte(`server { }`)),
			expected: &Config{
				Server: &serverConfig{
					ListenAddress: "0.0.0.0",
					ListenPort:    8085,
					SocketPath:    defaultSocketPath,
					LogLevel:      "INFO",
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
			err:    "unable to decode configuration: At 1:17: key 'not a valid hcl' expected start of object ('{') or assignment ('=')",
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
				assert.EqualError(t, err, tt.err)
				return
			}

			assert.Equal(t, tt.expected, serverConfig)
			assert.NoError(t, err)
		})
	}
}
