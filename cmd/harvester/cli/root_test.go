package cli

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewRootCmd(t *testing.T) {
	expected := &cobra.Command{
		Use:  "harvester",
		Long: "This is Galadriel Harvester CLI",
	}
	assert.Equal(t, expected, NewRootCmd())
}

func TestRootCmd(t *testing.T) {
	expectedSuccess := 0
	expectedError := 1
	cmdExecute = func() error {
		return nil
	}

	assert.Equal(t, Execute(), expectedSuccess)

	cmdExecute = func() error {
		return errors.New("")
	}

	assert.Equal(t, Execute(), expectedError)
}
