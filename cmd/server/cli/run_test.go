package cli

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	tempFile, err := os.CreateTemp("", "server.conf")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(`
		server {
			listen_address = "localhost"
			listen_port = "8085"
			socket_path = "/tmp/galadriel-server/api.sock/"
			log_level = "DEBUG"
}
		providers {
			Datastore "postgres" {
			connection_string = "postgresql://postgres:postgres@localhost:5432/galadriel"
			}

			X509CA "disk" {
				key_file_path = "./conf/server/dummy_root_ca.key"
				cert_file_path = "./conf/server/dummy_root_ca.crt"
			}

			KeyManager "memory" {}
}
	`)
	assert.NoError(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("config", tempFile.Name(), "")
	cmd.Flags().String("socketPath", "/test.api", "")

	config, err := LoadConfig(cmd)

	assert.NoError(t, err)
	assert.NotNil(t, config)
}
