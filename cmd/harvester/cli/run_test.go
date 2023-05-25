package cli

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary configuration file
	tempFile, err := os.CreateTemp("", "harvester.conf")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write sample configuration data to the temporary file
	_, err = tempFile.WriteString(`
		harvester {
	trust_domain = "example.org"
    harvester_socket_path = "/tmp/galadriel-harvester/api.sock"
    spire_socket_path = "/tmp/api.sock"
    galadriel_server_address = "localhost:5000"
    server_trust_bundle_path = "./root_ca.crt"
    federated_bundles_poll_interval = "1h"
	spire_bundle_poll_interval = "30m"
    log_level = "DEBUG"
	data_dir = "/test"
}

providers {
     BundleSigner "noop" {}
     BundleVerifier "noop" {}
}
	`)
	assert.NoError(t, err)

	// Create a mock command with the required flags
	cmd := &cobra.Command{}
	cmd.Flags().String("socketPath", "/test.api", "")
	cmd.Flags().String("config", tempFile.Name(), "")
	cmd.Flags().String("joinToken", "abc123", "")

	config, err := LoadConfig(cmd)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "/test.api", config.HarvesterSocketPath.String())
	assert.Equal(t, "abc123", config.JoinToken)
}
