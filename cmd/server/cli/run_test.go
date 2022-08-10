package cli

import (
	"fmt"
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

	if err != nil {
		fmt.Printf("%s", err)
	}

	assert.Equal(t, called, true)
}
