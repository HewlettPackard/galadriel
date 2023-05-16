package cli

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary configuration file
	tempFile, err := os.CreateTemp("", "harvester.conf")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write sample configuration data to the temporary file
	_, err = tempFile.WriteString(`
		harvester {
    local_socket_path = "/tmp/galadriel-harvester/api.sock"
    spire_socket_path = "/tmp/api.sock"
    server_address = "localhost:5000"
    server_trust_bundle_path = "./root_ca.crt"
    bundle_updates_interval = "1h"
    log_level = "DEBUG"
}
	`)
	assert.NoError(t, err)

	// Create a mock command with the required flags
	cmd := &cobra.Command{}
	cmd.Flags().String("config", tempFile.Name(), "")
	cmd.Flags().String("joinToken", "abc123", "")

	config, err := LoadConfig(cmd)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "abc123", config.JoinToken)
}
