package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestTokenCmd(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("trustDomain", "test.com", "")

	err := cmd.Execute()

	assert.Nil(t, err)
}
