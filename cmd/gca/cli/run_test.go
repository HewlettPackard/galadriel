package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("testdata/ok.conf")
	if err != nil {
		return
	}

	assert.Equal(t, "127.0.0.1:7000", config.TCPAddress.String())
	assert.Equal(t, "/socket-path", config.LocalAddress.String())
	assert.Equal(t, "root-cert", config.RootCertPath)
	assert.Equal(t, "root-key", config.RootKeyPath)
}
