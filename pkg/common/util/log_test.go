package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogSanitize(t *testing.T) {

	t.Run("Should clear non well formatted log message", func(t *testing.T) {
		msg := `
			Extensive messages 
			with line breakers should 
			be sanitized to avoid 
			log injection
		`

		sanitizedMsg := LogSanitize(msg)

		assert.NotEqual(t, sanitizedMsg, msg)
		assert.NotContains(t, sanitizedMsg, "\n")
		assert.NotContains(t, sanitizedMsg, "\r")
		t.Logf("Lined: %v", sanitizedMsg)
	})
}
