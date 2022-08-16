package telemetry

import (
	"context"
	"fmt"
	"net/http"

	"github.com/HewlettPackard/galadriel/pkg/common"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

type MetricServer interface {
	common.RunnablePlugin
}

type LocalMetricServer struct {
	logger common.Logger
}

func NewLocalMetricServer() MetricServer {
	return &LocalMetricServer{
		logger: *common.NewLogger(MetricsServer),
	}
}

func (c *LocalMetricServer) Run(ctx context.Context) error {
	c.logger.Info("Starting metric server")

	if err := runtimemetrics.Start(); err != nil {
		panic(err)
	}

	// TODO: verify telemetry plugin config to run specified plugin
	exporter := configurePrometheusMetrics()
	http.HandleFunc("/metrics", exporter.ServeHTTP)

	go func() {
		port := "8888"
		c.logger.Info("Listening on http://localhost:8888/metrics")

		address := fmt.Sprintf(":%s", port)
		err := http.ListenAndServe(address, nil)
		if err != nil {
			panic(err)
		}
	}()

	<-ctx.Done()
	return nil
}

func configurePrometheusMetrics() *prometheus.Exporter {
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

	exporter, err := prometheus.New(config, ctrl)
	if err != nil {
		panic(err)
	}

	global.SetMeterProvider(exporter.MeterProvider())

	return exporter
}

func FormatLabel(component, entity, action string) string {
	return fmt.Sprintf("%s.%s.%s", component, entity, action)
}

func InitializeMeterProvider(ctx context.Context) metric.Meter {
	type key string
	ctxKey := key(PackageName)
	packageName := fmt.Sprintf("%v", ctx.Value(ctxKey))

	return global.MeterProvider().Meter(packageName)
}
