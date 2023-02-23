package telemetry

import (
	"context"
	"errors"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/armon/go-metrics"
)

// Label is a label/tag for a metric
type Label = metrics.Label

// Sink is an interface for emitting metrics
type Sink = metrics.MetricSink

// Metrics is an interface for all metrics plugins and services
type Metrics interface {
	// A Gauge should retain the last value it is set to
	SetGauge(key []string, val float32)
	SetGaugeWithLabels(key []string, val float32, labels []Label)

	// Should emit a Key/Value pair for each call
	EmitKey(key []string, val float32)

	// Counters should accumulate values
	IncrCounter(key []string, val float32)
	IncrCounterWithLabels(key []string, val float32, labels []Label)

	// Samples are for timing information, where quantiles are used
	AddSample(key []string, val float32)
	AddSampleWithLabels(key []string, val float32, labels []Label)

	// A convenience function for measuring elapsed time with a single line
	MeasureSince(key []string, start time.Time)
	MeasureSinceWithLabels(key []string, start time.Time, labels []Label)
}

type MetricsImpl struct {
	*metrics.Metrics

	c            *MetricsConfig
	runners      []sinkRunner
	metricsSinks []*metrics.Metrics
}

// NewMetrics returns a Metric implementation
func NewMetrics(c *MetricsConfig) (*MetricsImpl, error) {
	if c.Logger == nil {
		return nil, errors.New("logger must be configured")
	}

	metricsImpl := &MetricsImpl{c: c}

	// loop to support multiple metrics collectors in the future
	for _, f := range sinkRunnerFactories {
		runner, err := f(c)
		if err != nil {
			return nil, err
		}

		if !runner.isConfigured() {
			continue
		}

		fanout := metrics.FanoutSink{}
		fanout = append(fanout, runner.sinks()...)

		conf := metrics.DefaultConfig(c.ServiceName)
		conf.EnableHostname = false
		conf.EnableHostnameLabel = true

		metricsSink, err := metrics.New(conf, fanout)
		if err != nil {
			return nil, err
		}

		metricsImpl.metricsSinks = append(metricsImpl.metricsSinks, metricsSink)
		metricsImpl.runners = append(metricsImpl.runners, runner)
	}

	return metricsImpl, nil
}

// ListenAndServe starts the metrics process
func (m *MetricsImpl) ListenAndServe(ctx context.Context) error {
	var tasks []util.RunnableTask
	for _, runner := range m.runners {
		tasks = append(tasks, runner.run)
	}

	return util.RunTasks(ctx, tasks...)
}

// SetGauge sets the gauge for the metrics
func (m *MetricsImpl) SetGauge(key []string, val float32) {
	for _, s := range m.metricsSinks {
		s.SetGauge(key, val)
	}
}

// SetGaugeWithLabels sets the gauge with labels/tags
func (m *MetricsImpl) SetGaugeWithLabels(key []string, val float32, labels []Label) {
	for _, s := range m.metricsSinks {
		s.SetGaugeWithLabels(key, val, labels)
	}
}

func (m *MetricsImpl) EmitKey(key []string, val float32) {
	for _, s := range m.metricsSinks {
		s.EmitKey(key, val)
	}
}

// IncrCounter increments the counter for the given metric.
func (m *MetricsImpl) IncrCounter(key []string, val float32) {
	for _, s := range m.metricsSinks {
		s.IncrCounter(key, val)
	}
}

// IncrCounterWithLabels delegates to embedded metrics
func (m *MetricsImpl) IncrCounterWithLabels(key []string, val float32, labels []Label) {
	for _, s := range m.metricsSinks {
		s.IncrCounterWithLabels(key, val, labels)
	}
}

// AddSample adds a sample to the metrics
func (m *MetricsImpl) AddSample(key []string, val float32) {
	for _, s := range m.metricsSinks {
		s.AddSample(key, val)
	}
}

// AddSampleWithLabels add a sample value to the metrics with labels/tag
func (m *MetricsImpl) AddSampleWithLabels(key []string, val float32, labels []Label) {
	for _, s := range m.metricsSinks {
		s.AddSampleWithLabels(key, val, labels)
	}
}

// MeasureSince measure the elapsed time
func (m *MetricsImpl) MeasureSince(key []string, start time.Time) {
	for _, s := range m.metricsSinks {
		s.MeasureSince(key, start)
	}
}

// MeasureSinceWithLabels measure the elapsed time with labels/tags embedded to it
func (m *MetricsImpl) MeasureSinceWithLabels(key []string, start time.Time, labels []Label) {
	for _, s := range m.metricsSinks {
		s.MeasureSinceWithLabels(key, start, labels)
	}
}
