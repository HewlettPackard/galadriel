package catalog

import (
	"context"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/sirupsen/logrus"
)

type Catalog interface {
	// methods for accessing pluggable interfaces (e.g., NodeAttestors)
	// no needed for a PoC.
}

type Repository struct {
	logger logrus.FieldLogger
}

func (r *Repository) Close() {
	// TODO: close repository
}

type Config struct {
	Logger  logrus.FieldLogger
	Metrics telemetry.MetricServer
}

func Load(ctx context.Context, config Config) (*Repository, error) {
	re := &Repository{
		logger: config.Logger,
	}

	return re, nil
}
