package telemetry

import (
	"context"
	"fmt"
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
	logger common.Logger
}

func NewLocalMetricServer() MetricServer {
	return &LocalMetricServer{
		logger: *common.NewLogger("metric_server"),
	}
}

func (c *LocalMetricServer) Run(ctx context.Context) error {
	c.logger.Info("Starting metric server")

	exporter := configureMetrics()

	if err := runtimemetrics.Start(); err != nil {
		panic(err)
	}

	http.HandleFunc("/metrics", exporter.ServeHTTP)
	fmt.Println("listenening on http://localhost:8088/metrics")

	go func() {
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
