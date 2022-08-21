package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TesRun(t *testing.T) {
	expectedSuccess := 0
	expectedError := 1
	cmdExecute = func() error {
		return nil
	}

	assert.Equal(t, expectedSuccess, Execute())

	cmdExecute = func() error {
		return errors.New("Ops")
	}

	assert.Equal(t, expectedError, Execute())
}
