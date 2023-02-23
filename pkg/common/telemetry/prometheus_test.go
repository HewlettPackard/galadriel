package telemetry

import (
	"context"
	"errors"
	"testing"
	"time"

	prommetrics "github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errNotConfigured = errors.New("configuration file not specified")
)

func TestNewPrometheusRunner(t *testing.T) {
	config := testPrometheusConfig()

	t.Run("Configuration is properly set and no error is returned", func(t *testing.T) {
		pr, err := newTestPrometheusRunner(config)
		assert.Nil(t, err)
		assert.NotNil(t, pr)
	})

	t.Run("No error is returned but the configuration is nil", func(t *testing.T) {
		config.PrometheusConf = nil
		pr, err := newTestPrometheusRunner(config)
		assert.NotNil(t, err)
		assert.Nil(t, pr)
		assert.Equal(t, err, errNotConfigured)
	})

}

func TestConfiguration(t *testing.T) {
	config := testPrometheusConfig()

	t.Run("Success when the config is properly filled", func(t *testing.T) {
		pr, err := newTestPrometheusRunner(config)
		require.NoError(t, err)
		assert.True(t, pr.isConfigured())
	})

	t.Run("Error when the config is missing required properties", func(t *testing.T) {
		config.PrometheusConf = nil
		pr, err := newTestPrometheusRunner(config)
		assert.NotNil(t, err)
		assert.Nil(t, pr)
		assert.Equal(t, err, errNotConfigured)
	})
}

func TestRun(t *testing.T) {
	config := testPrometheusConfig()
	errCh := make(chan error)

	t.Run("Runs and stops when the context is canceled", func(t *testing.T) {
		pr, err := newTestPrometheusRunner(config)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			errCh <- pr.run(ctx)
		}()

		cancel()
		select {
		case err := <-errCh:
			assert.Equal(t, context.Canceled, err)
		case <-time.After(time.Minute):
			t.Fatal("timeout waiting for shutdown")
		}

	})

	t.Run("Does not run if not configured", func(t *testing.T) {
		config.PrometheusConf = nil
		pr, err := newTestPrometheusRunner(config)
		assert.Nil(t, pr)
		assert.NotNil(t, err)
		assert.Equal(t, err, errNotConfigured)
	})

}

func testPrometheusConfig() *MetricsConfig {
	l, _ := test.NewNullLogger()

	return &MetricsConfig{
		Logger:         l,
		ServiceName:    "foo",
		PrometheusConf: &PrometheusConfig{},
	}
}

// newTestPrometheusRunner wraps newPrometheusRunner, unregistering the
// collector after creation in order to avoid duplicate registration errors
func newTestPrometheusRunner(c *MetricsConfig) (sinkRunner, error) {
	runner, err := newPrometheusRunner(c)

	if runner != nil && runner.isConfigured() {
		pr := runner.(*prometheusRunner)
		sink := pr.sink.(*prommetrics.PrometheusSink)
		prometheus.Unregister(sink)
	}

	return runner, err
}
