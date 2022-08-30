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
		expected *Server
		err      string
	}{
		{
			name:   "ok",
			config: bytes.NewBuffer([]byte(`server { listen_address = "listen_address" }`)),
			expected: &Server{
				ServerConfigSection: &ServerConfigSection{
					ListenAddress: "listen_address",
					LogLevel:      "INFO",
				},
			},
		},
		{
			name:   "defaults",
			config: bytes.NewBuffer([]byte(`server { }`)),
			expected: &Server{
				ServerConfigSection: &ServerConfigSection{
					ListenAddress: "localhost:8080",
					LogLevel:      "INFO",
				},
			},
		},
		{
			name:   "empty_config_file",
			config: bytes.NewBufferString(``),
			err:    "bad configuration: configuration file is empty",
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
			serverConfig, err := New(tt.config)

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
