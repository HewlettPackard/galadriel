package telemetry

import (
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

func TestNewLocalMetricServer(t *testing.T) {
	expected := &LocalMetricServer{
		logger: *common.NewLogger(MetricsServer),
	}

	metricServer := NewLocalMetricServer()
	assert.Equal(t, expected, metricServer)
}

func TestFormatLabel(t *testing.T) {
	component := "component"
	entity := "entity"
	action := "action"
	expected := "component.entity.action"
	label := FormatLabel(component, entity, action)

	assert.Equal(t, expected, label)
}

func TestConfigurePrometheusMetrics(t *testing.T) {
	config := prometheus.Config{}
	ctrl := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
	)

	expected, err := prometheus.New(config, ctrl)

	exporter := configurePrometheusMetrics()
	assert.ObjectsAreEqual(expected, exporter)
	assert.Nil(t, err)
}
