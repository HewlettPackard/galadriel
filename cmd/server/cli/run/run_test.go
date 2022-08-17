package run

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCommand(t *testing.T) {
	called := false

	runServerFn = func() {
		called = true
	}

	cmd := NewRunCmd()
	err := cmd.Execute()

	assert.Equal(t, err, nil)
	assert.Equal(t, called, true)
}
