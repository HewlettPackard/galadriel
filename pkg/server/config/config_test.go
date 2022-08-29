package config

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeReader int

func (fakeReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("error from fake reader")
}

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		config   io.Reader
		expected *ServerConfig
		err      string
	}{
		{
			name:   "ok",
			config: bytes.NewBuffer([]byte(`server { spire_socket_path = "spire_socket_path" server_address = "server_address" }`)),
			expected: &ServerConfig{
				ServerConfigSection: &ServerConfigSection{
					SpireSocketPath: "spire_socket_path",
					ServerAddress:   "server_address",
					LogLevel:        "INFO",
				},
			},
		},
		{
			name:   "defaults",
			config: bytes.NewBuffer([]byte(`server { server_address = "server_address" }`)),
			expected: &ServerConfig{
				ServerConfigSection: &ServerConfigSection{
					SpireSocketPath: "/tmp/spire-server/private/api.sock",
					ServerAddress:   "server_address",
					LogLevel:        "INFO",
				},
			},
		},
		{
			name:   "empty_config_file",
			config: bytes.NewBufferString(``),
			err:    "invalid configuration: server.server_address is required",
		},
		{
			name:   "requires_server_address",
			config: bytes.NewBufferString(`server { spire_socket_path = "test" }`),
			err:    "invalid configuration: server.server_address is required",
		},
		{
			name:   "invalid_hcl",
			config: bytes.NewBufferString(`not a valid hcl`),
			err:    "unable to decode configuration: At 1:17: key 'not a valid hcl' expected start of object ('{') or assignment ('=')",
		},
		{
			name:   "invalid_config_reader",
			config: nil,
			err:    "configuration is required",
		},
		{
			name:   "invalid_config_reader_error",
			config: fakeReader(0),
			err:    "failed to read configuration: error from fake reader",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.config)

			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}

}
