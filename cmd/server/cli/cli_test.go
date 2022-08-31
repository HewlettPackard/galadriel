package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLI(t *testing.T) {
	expectedSuccess := 0
	expectedError := 1
	cmdExecute = func() error {
		return nil
	}

	assert.Equal(t, expectedSuccess, Run())

	cmdExecute = func() error {
		return errors.New("")
	}

	assert.Equal(t, expectedError, Run())
}
