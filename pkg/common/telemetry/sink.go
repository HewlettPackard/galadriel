package telemetry

import "context"

// TODO: support multiple metrics collectors for the application.
var sinkRunnerFactories = []sinkRunnerFactory{
	newPrometheusRunner,
}

type sinkRunner interface {
	isConfigured() bool
	sinks() []Sink

	run(context.Context) error
}

type sinkRunnerFactory func(*MetricsConfig) (sinkRunner, error)
