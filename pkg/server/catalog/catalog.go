package catalog

import (
	"context"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/sirupsen/logrus"
)

type Catalog interface {
	GetDataStore() datastore.DataStore
	// methods for accessing pluggable interfaces (e.g., attestors, upstreamCAs, etc)
}

type Repository struct {
	DataStore datastore.DataStore
	log       logrus.FieldLogger
}

func (r *Repository) GetDataStore() datastore.DataStore {
	return r.DataStore
}

func (r *Repository) Close() {
	// TODO: close repository
}

type Config struct {
	Log     logrus.FieldLogger
	Metrics telemetry.MetricServer
}

func Load(ctx context.Context, config Config) (*Repository, error) {
	memStore := datastore.MemStore{}

	re := &Repository{
		DataStore: &memStore,
		log:       logrus.StandardLogger(),
	}

	return re, nil
}
