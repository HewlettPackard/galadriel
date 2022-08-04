package telemetry

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		config   io.Reader
		expected *TelemetryConfig
		err      string
	}{
		{
			name:   "ok",
			config: bytes.NewBuffer([]byte(`telemetry { Prometheus { host = "localhost" port = "9090"} }`)),
			expected: &TelemetryConfig{
				TelemetryConfigSection: &TelemetryConfigSection{
					&PrometheusConfig{
						Host: "localhost",
						Port: 9090,
					},
				},
			},
		},
		{
			name:   "invalid_hcl",
			config: bytes.NewBufferString(`not a valid hcl`),
			err:    "unable to decode configuration: At 1:17: key 'not a valid hcl' expected start of object ('{') or assignment ('=')",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.config)

			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}

}
