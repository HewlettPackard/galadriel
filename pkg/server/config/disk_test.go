package config

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFromDisk(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected *Server
		err      string
	}{
		{
			name: "ok",
			path: "testdata/ok.conf",
			expected: &Server{
				ServerConfigSection: &ServerConfigSection{
					SpireSocketPath: "spire_socket_path",
					ServerAddress:   "server_address",
					LogLevel:        "log_level",
				},
			},
		}, {
			name:     "ok_empty",
			path:     "testdata/empty.conf",
			expected: &Server{},
		},
		{
			name:     "cannot_open_file",
			path:     "invalid_path",
			expected: nil,
			err:      "unable to open configuration file: open invalid_path: no such file or directory",
		},
		{
			name:     "new_error",
			path:     "testdata/empty.conf",
			expected: nil,
			err:      "unable to parse configuration file: unexpected end of JSON input",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			newFn = func(config io.Reader) (*Server, error) {
				if tt.err != "" {
					return nil, errors.New(tt.err)
				}
				return tt.expected, nil
			}

			got, err := LoadFromDisk(tt.path)

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
