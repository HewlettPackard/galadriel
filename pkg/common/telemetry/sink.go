package telemetry

import "context"

var sinkRunnerFactories = []sinkRunnerFactory{
	newPrometheusRunner,
}

type sinkRunner interface {
	isConfigured() bool
	sinks() []Sink

	run(context.Context) error
}

type sinkRunnerFactory func(*MetricsConfig) (sinkRunner, error)
