package cli

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	cmdExecute = func() error {
		return nil
	}
	assert.Nil(t, HarvesterCLI.Execute())

	cmdExecute = func() error {
		return errors.New("Ops")
	}
	assert.NotNil(t, HarvesterCLI.Execute())
}

func TestNewRootCmd(t *testing.T) {
	expected := &cobra.Command{
		Use:  "harvester",
		Long: "This is the Galadriel Harvester CLI",
	}
	assert.Equal(t, expected, NewRootCmd())
}
