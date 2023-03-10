package cli

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type fakeReader int

func (fakeReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("error from fake reader")
}

func TestNewGCAConfig(t *testing.T) {
	config := Config{GCA: &gcaConfig{
		ListenAddress: "localhost",
		ListenPort:    8000,
		SocketPath:    "/example",
		LogLevel:      "INFO",
		RootCertPath:  "root_cert_path",
		RootKeyPath:   "rootKey_path",
		X509CertTTL:   "5h",
		JWTTokenTTL:   "10h",
	}}

	sc, err := NewGCAConfig(&config)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "127.0.0.1", sc.TCPAddress.IP.String())
	assert.Equal(t, config.GCA.ListenPort, sc.TCPAddress.Port)
	assert.Equal(t, config.GCA.SocketPath, sc.LocalAddress.String())
	assert.Equal(t, strings.ToLower(config.GCA.LogLevel), logrus.GetLevel().String())
	assert.Equal(t, config.GCA.RootCertPath, sc.RootCertPath)
	assert.Equal(t, config.GCA.RootKeyPath, sc.RootKeyPath)
	assert.Equal(t, time.Duration(18000000000000), sc.X509CertTTL)
	assert.Equal(t, time.Duration(36000000000000), sc.JWTCertTTL)
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
				`gca { listen_address = "127.0.0.1" listen_port = 2222 socket_path = "/tmp/api.sock" log_level = "INFO" 
root_cert_path = "root-cert" root_key_path = "root-key"  x509_cert_ttl = "2h" jwt_token_ttl = "5h"}`)),
			expected: &Config{
				GCA: &gcaConfig{
					ListenAddress: "127.0.0.1",
					ListenPort:    2222,
					SocketPath:    "/tmp/api.sock",
					LogLevel:      "INFO",
					RootCertPath:  "root-cert",
					RootKeyPath:   "root-key",
					X509CertTTL:   "2h",
					JWTTokenTTL:   "5h",
				},
			},
		},
		{
			name:   "defaults",
			config: bytes.NewBuffer([]byte(`gca { }`)),
			expected: &Config{
				GCA: &gcaConfig{
					ListenAddress: defaultAddress,
					ListenPort:    defaultPort,
					SocketPath:    defaultSocketPath,
					LogLevel:      defaultLogLevel,
					X509CertTTL:   defaultX509CertTTL,
					JWTTokenTTL:   defaultJWTTokenTTL,
				},
			},
		},
		{
			name:   "empty_config_file",
			config: bytes.NewBufferString(``),
			err:    "gca section is empty",
		},
		{
			name:   "err_hcl",
			config: bytes.NewBufferString(`not a valid hcl`),
			err:    "failed to decode configuration: At 1:17: key 'not a valid hcl' expected start of object ('{') or assignment ('=')",
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
