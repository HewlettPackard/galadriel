package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCmd(t *testing.T) {
	cmdExecute = func() error {
		return nil
	}

	assert.Equal(t, Execute(), 0)

	cmdExecute = func() error {
		return errors.New("")
	}

	assert.Equal(t, Execute(), 1)
}
