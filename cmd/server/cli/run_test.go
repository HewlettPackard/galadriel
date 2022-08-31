package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCommand(t *testing.T) {
	called := false
	runServerFn = func(config string) error {
		called = true
		return nil
	}

	cmd := NewRunCmd()
	cmd.Flags().StringVarP(&configPath, "config", "c", defaultConfigPath, "config file path")
	err := cmd.Execute()

	assert.Equal(t, nil, err, "unexpected error")
	assert.Equal(t, true, called, "failed to call runServerFn")

	called = false
	runServerFn = func(config string) error {
		called = true
		return errors.New("ops")
	}

	err = cmd.Execute()

	assert.Equal(t, errors.New("ops"), err, "unexpected error")
	assert.Equal(t, true, called, "failed to call runServerFn")
}
