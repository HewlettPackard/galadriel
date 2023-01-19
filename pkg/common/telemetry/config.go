package telemetry

import "github.com/sirupsen/logrus"

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Logger logrus.FieldLogger

	// ServiceName is the name of the service that is being monitored
	ServiceName string

	// Sink represents the interface for sending metrics
	Sinks []Sink

	// PrometheusConf conveys the configuration for Prometheus
	PrometheusConf *PrometheusConfig `hcl:"Prometheus"`
}

// PrometheusConfig represents the Prometheus configuration
type PrometheusConfig struct {
	// Host is the Prometheus server host eg: "localhost"
	Host string `hcl:"host"`

	//Port is the Prometheus server port eg:"9000"
	Port int `hcl:"port"`
}
