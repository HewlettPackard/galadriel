package telemetry

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/HewlettPackard/Galadriel/pkg/common"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
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
	logger      common.Logger
	FileConfig  FileConfig
	ServiceName string
}

type FileConfig struct {
	Prometheus *PrometheusConfig `hcl:"Prometheus"`
}

type PrometheusConfig struct {
	Host string `hcl:"host"`
	Port int    `hcl:"port"`
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

	exporter := configureMetrics()
	http.HandleFunc("/metrics", exporter.ServeHTTP)

	go func() {
		c.logger.Info("listenening on http://localhost:8088/metrics")
		_ = http.ListenAndServe(":8088", nil)
	}()

	sampleMetric(ctx)

	<-ctx.Done()
	return nil
}

func configureMetrics() *prometheus.Exporter {
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

func sampleMetric(ctx context.Context) {
	meter := global.MeterProvider().Meter("example")
	counter, err := meter.SyncInt64().Counter(
		"test.my_counter",
		instrument.WithDescription("Just a test counter"),
	)
	if err != nil {
		panic(err)
	}

	for {
		n := rand.Intn(100)
		time.Sleep(time.Duration(n) * time.Millisecond)

		counter.Add(ctx, 1)
	}
}
