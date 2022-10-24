package catalog

import (
	"context"

	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/sirupsen/logrus"
)

type Catalog interface {
	GetDataStore() datastore.DataStore
	// methods for accessing pluggable interfaces (e.g., attestors, upstreamCAs, etc)
}

type Repository struct {
	DataStore datastore.DataStore
	logger    logrus.FieldLogger
}

func (r *Repository) GetDataStore() datastore.DataStore {
	return r.DataStore
}

func (r *Repository) Close() {
	// TODO: close repository
}

type Config struct {
	Logger logrus.FieldLogger
}

func Load(ctx context.Context, config Config) (*Repository, error) {
	re := &Repository{
		DataStore: datastore.NewMemStore(),
		logger:    config.Logger,
	}

	return re, nil
}
